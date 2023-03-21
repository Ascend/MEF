// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package websocketmgr for websocket manager
package websocketmgr

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"huawei.com/mindx/common/hwlog"
)

type wsMessage struct {
	MsgType int
	Value   []byte
}

// WsClientProxy websocket client proxy
type WsClientProxy struct {
	ProxyCfg *ProxyConfig
	connMgr  *wsConnectMgr
}

// waiting a moment for edgeproxy server is ready
const delayStartTime = time.Second * 3

// GetName get websocket client name
func (wcp *WsClientProxy) GetName() string {
	return wcp.ProxyCfg.name
}

// Start websocket client start
func (wcp *WsClientProxy) Start() error {
	if err := wcp.start(); err != nil {
		hwlog.RunLog.Errorf("websocket start failed, error: %v", err)
		return fmt.Errorf("websocket cilent proxy start falied, error: %v", err)
	}
	go wcp.ReConnect()
	return nil
}

func (wcp *WsClientProxy) start() error {
	dialer := &websocket.Dialer{
		TLSClientConfig:  wcp.ProxyCfg.tlsConfig,
		HandshakeTimeout: handshakeTimeout,
		ReadBufferSize:   readBufferSize,
		WriteBufferSize:  writeBufferSize,
	}
	hwlog.RunLog.Info("websocket client begin try to connect the server")
	connect, err := wcp.tryConnect(dialer)
	if err != nil {
		return fmt.Errorf("connect the server failed, error: %v", err)
	}
	wcp.connMgr = &wsConnectMgr{}
	wcp.connMgr.start(connect, wcp.ProxyCfg.name, &wcp.ProxyCfg.handlerMgr)
	return nil
}

// ReConnect  reconnecting with websocket server
func (wcp *WsClientProxy) ReConnect() {
	for {
		select {
		case <-wcp.ProxyCfg.ctx.Done():
			return
		default:
		}
		time.Sleep(reconnectWaitTime)
		if wcp.IsConnected() {
			continue
		}
		hwlog.RunLog.Info("websocket client start reconnecting now...")
		if err := wcp.start(); err != nil {
			hwlog.RunLog.Errorf("websocket client start failed, error: %v", err)
			continue
		}
	}
}

// Stop websocket client stop
func (wcp *WsClientProxy) Stop() error {
	wcp.ProxyCfg.cancel()
	err := wcp.connMgr.stop()
	if err != nil {
		return fmt.Errorf("stop websocket connection failed, error: %v", err)
	}
	return nil
}

// Send websocket client send message
func (wcp *WsClientProxy) Send(msg interface{}) error {
	if !wcp.connMgr.isConnected() {
		return fmt.Errorf("websocket not connect, please connect first")
	}

	wsMsg, ok := msg.(wsMessage)
	if !ok {
		return fmt.Errorf("websocket client send message failed, the message type unsupported")
	}

	err := wcp.connMgr.send(wsMsg)
	if err != nil {
		return err
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
		if errCnt += 1; errCnt%3 == 0 {
			hwlog.RunLog.Errorf("websocket client connect the server failed, error: %v", err)
		}
		time.Sleep(tryConnInterval)
		if tryConnInterval < maxTryConnInterval {
			tryConnInterval = tryConnInterval * tryConnIntervalGrowRate
		}
	}
}
