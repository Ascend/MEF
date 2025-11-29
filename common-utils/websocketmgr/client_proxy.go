// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package websocketmgr for websocket manager
package websocketmgr

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/gorilla/websocket"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/limiter"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/utils"
)

// WsClientProxy websocket client proxy
type WsClientProxy struct {
	ProxyCfg            *ProxyConfig
	connMgr             *wsConnectMgr
	reconnectCallbacks  []func()
	disconnectCallbacks []func(WebsocketPeerInfo)
	bandwidthLimiter    limiter.ServerBandwidthLimiterIntf
}

// waiting a moment for server is ready
const delayStartTime = time.Second * 3
const errorPrintFrequency = 3

// GetName get websocket client name
func (wcp *WsClientProxy) GetName() string {
	return wcp.ProxyCfg.name
}

// Start websocket client start
func (wcp *WsClientProxy) Start() error {
	wcp.ProxyCfg.ctx, wcp.ProxyCfg.cancel = context.WithCancel(context.Background())
	if err := wcp.start(); err != nil {
		hwlog.RunLog.Errorf("websocket start failed, error: %v", err)
		return fmt.Errorf("websocket cilent proxy start falied, error: %v", err)
	}
	go wcp.reconnect()
	return nil
}

func (wcp *WsClientProxy) start() error {
	dialer := &websocket.Dialer{
		TLSClientConfig:  wcp.ProxyCfg.tlsConfig,
		HandshakeTimeout: defaultHandshakeTimeout,
		ReadBufferSize:   wsReadBufferSize,
		WriteBufferSize:  wsWriteBufferSize,
	}
	hwlog.RunLog.Info("websocket client begin try to connect the server")
	conn, err := wcp.tryConnect(dialer)
	if err != nil {
		return fmt.Errorf("connect the server failed, error: %v", err)
	}
	host, _, err := net.SplitHostPort(wcp.ProxyCfg.hosts)
	if err != nil {
		return fmt.Errorf("get host ip failed: %v", err)
	}

	// initialize then start bandwidth limiter immediately when client proxy starts
	bandwidthLimiter := limiter.NewServerBandwidthLimiter(wcp.ProxyCfg.bandwidthLimiterCfg)
	if bandwidthLimiter != nil {
		if bwl := wcp.bandwidthLimiter; bwl != nil {
			bwl.Stop()
		}
		wcp.bandwidthLimiter = bandwidthLimiter
	}

	wcp.connMgr = &wsConnectMgr{
		conn:         conn,
		peerInfo:     WebsocketPeerInfo{Ip: host, Sn: wcp.ProxyCfg.name},
		currentProxy: wcp,
	}
	wcp.connMgr.start()
	return nil
}

func (wcp *WsClientProxy) reconnect() {
	for {
		select {
		case <-wcp.ProxyCfg.ctx.Done():
			hwlog.RunLog.Info("Stop reconnect")
			return
		default:
		}
		time.Sleep(reconnectInterval)
		if wcp.IsConnected() {
			continue
		}
		hwlog.RunLog.Info("websocket client start reconnecting now...")
		if err := wcp.start(); err != nil {
			hwlog.RunLog.Errorf("websocket client start failed, error: %v", err)
			continue
		}

		// execute the callback function after reconnection
		for _, callback := range wcp.reconnectCallbacks {
			callback()
		}
	}
}

// Stop websocket connection
func (wcp *WsClientProxy) Stop() error {
	wcp.ProxyCfg.cancel()
	// deactivate all message handler limiter when client proxy stops
	if wcp.bandwidthLimiter != nil {
		wcp.bandwidthLimiter.Stop()
	}
	err := wcp.connMgr.stop()
	if err != nil {
		return fmt.Errorf("stop websocket connection failed, error: %v",
			utils.TrimInfoFromError(err))
	}
	return nil
}

// Send websocket client send message
func (wcp *WsClientProxy) Send(msg *model.Message, msgType ...int) error {
	wsMsgType := websocket.TextMessage
	if len(msgType) > 0 {
		wsMsgType = msgType[0]
	}
	if !isValidWsMsgType(wsMsgType) {
		return fmt.Errorf("websocket message type [%v] is not supported", wsMsgType)
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message failed: %v", err)
	}
	if err := wcp.connMgr.send(wsMsgType, data); err != nil {
		return utils.TrimInfoFromError(err)
	}
	return nil
}

// IsConnected judge the client is connected
func (wcp *WsClientProxy) IsConnected() bool {
	return wcp.connMgr.isConnected()
}

func (wcp *WsClientProxy) tryConnect(dialer *websocket.Dialer) (*websocket.Conn, error) {
	tryConnInterval := 1 * time.Second
	errCnt := 0
	time.Sleep(delayStartTime)
	for {
		select {
		case <-wcp.ProxyCfg.ctx.Done():
			return nil, fmt.Errorf("connect has be canceled")
		default:
		}
		hwlog.RunLog.Info("websocket client try to connect server")
		connect, _, err := dialer.Dial(wssProtocol+wcp.ProxyCfg.hosts, wcp.ProxyCfg.headers)
		if err == nil {
			hwlog.RunLog.Info("websocket client connect the server success")
			return connect, nil
		}
		// print error msg each 3-times, not every time
		if errCnt += 1; errCnt%errorPrintFrequency == 0 {
			hwlog.RunLog.Errorf("websocket client connect the server failed, error: %v",
				utils.TrimInfoFromError(err))
		}
		time.Sleep(tryConnInterval)
		if tryConnInterval < maxTryConnInterval {
			tryConnInterval = tryConnInterval * tryConnIntervalGrowRate
		}
	}
}

// SetReConnCallback is the func to set the callback func when client disconnect
func (wcp *WsClientProxy) SetReConnCallback(cf ...func()) {
	if len(cf) == 0 {
		wcp.reconnectCallbacks = []func(){}
		return
	}
	wcp.reconnectCallbacks = cf
}

// SetDisconnCallback is the func to set the callback func when client disconnect
func (wcp *WsClientProxy) SetDisconnCallback(cf ...func(WebsocketPeerInfo)) {
	if len(cf) == 0 {
		wcp.disconnectCallbacks = []func(WebsocketPeerInfo){}
		return
	}
	wcp.disconnectCallbacks = cf
}

// GetReconnectCallbacks get all registered reconnect callback functions
func (wcp *WsClientProxy) GetReconnectCallbacks() []func() {
	return wcp.reconnectCallbacks
}

// GetDisconnectCallbacks get all registered disconnect callback functions
func (wcp *WsClientProxy) GetDisconnectCallbacks() []func(WebsocketPeerInfo) {
	return wcp.disconnectCallbacks
}

// GetOnConnectCallbacks get all registered on connection callback functions
func (wcp *WsClientProxy) GetOnConnectCallbacks() []func(WebsocketPeerInfo) {
	return nil
}

// GetBandwidthLimiter Get bandwidth limiter instance, it not set, return nil
func (wcp *WsClientProxy) GetBandwidthLimiter() limiter.ServerBandwidthLimiterIntf {
	return wcp.bandwidthLimiter
}

// GetProxyConfig Get proxy config for server proxy
func (wcp *WsClientProxy) GetProxyConfig() *ProxyConfig {
	return wcp.ProxyCfg
}
