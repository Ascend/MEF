// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package websocketmgr for websocket manager
package websocketmgr

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/limiter"
	"huawei.com/mindx/common/modulemgr/model"
)

// ProxyInstanceIntf get related proxy instance(eg: server proxy / client proxy) then retrieve callbacks and limiter
type ProxyInstanceIntf interface {
	GetReconnectCallbacks() []func()
	GetDisconnectCallbacks() []func(WebsocketPeerInfo)
	GetBandwidthLimiter() limiter.ServerBandwidthLimiterIntf
	GetProxyConfig() *ProxyConfig
	GetOnConnectCallbacks() []func(WebsocketPeerInfo)
}

// WebsocketPeerInfo contains client information
type WebsocketPeerInfo struct {
	Sn string
	Ip string
}

type wsConnectMgr struct {
	conn          *websocket.Conn
	peerInfo      WebsocketPeerInfo
	currentProxy  ProxyInstanceIntf
	lastAlive     time.Time
	connFlag      bool
	connLock      sync.RWMutex
	waitHandleGrp sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	sendLock      sync.Mutex
	rpsLimiter    limiter.IndependentLimiter // conn-level ws rps limiter, run before message-level limiter
}

const (
	pingMsg = "send ping message"
)

func (cm *wsConnectMgr) start() {
	cm.connLock.Lock()
	defer cm.connLock.Unlock()
	cm.ctx, cm.cancel = context.WithCancel(context.Background())
	cm.connFlag = true
	cm.lastAlive = time.Now()
	cm.conn.SetPongHandler(cm.pongHandle)
	cm.conn.SetPingHandler(cm.pingHandle)
	cm.conn.SetCloseHandler(cm.closeHandle)
	cm.conn.SetReadLimit(defaultReadSizeLimit)
	// initialize and start a connection related message limiter for each websocket connection
	if cfg := cm.currentProxy.GetProxyConfig().rpsLimiterCfg; cfg != nil {
		cm.rpsLimiter = limiter.NewRpsLimiter(cfg.Rps, cfg.Burst)
	}
	if bandwidthLimiter := cm.currentProxy.GetBandwidthLimiter(); bandwidthLimiter != nil {
		bandwidthLimiter.RegisterConn(cm.getPeerId())
	}
	if callbacks := cm.currentProxy.GetOnConnectCallbacks(); callbacks != nil {
		for _, cf := range callbacks {
			cf(cm.peerInfo)
		}
	}
	cm.startLoop()
}

func (cm *wsConnectMgr) getPeerId() string {
	return cm.peerInfo.Sn
}

func (cm *wsConnectMgr) isConnected() bool {
	return cm.connFlag && cm.conn != nil
}

func (cm *wsConnectMgr) startLoop() {
	go cm.heartbeat()
	go cm.receive()
}

func (cm *wsConnectMgr) receive() {
	cm.waitHandleGrp.Add(1)
	defer cm.postProcess()

	for {
		select {
		case <-cm.ctx.Done():
			hwlog.RunLog.Warnf("connect (%s) receive Stop", cm.getPeerId())
			cm.waitHandleGrp.Done()
			return
		default:
		}
		msg, err := cm.readMsg()
		if err != nil {
			hwlog.RunLog.Errorf("read next websocket message from [%v] error: %v", cm.getPeerId(), err)
			cm.waitHandleGrp.Done()
			if err := cm.stop(); err != nil {
				hwlog.RunLog.Errorf("stop websocket connection error: %v", err)
			}
			return
		}
		if msg == nil {
			continue
		}
		if err := cm.limiterCheck(len(msg)); err != nil {
			hwlog.RunLog.Warnf("%v", err)
			continue
		}
		peerInfo := model.MsgPeerInfo{Ip: cm.peerInfo.Ip, Sn: cm.peerInfo.Sn}
		handler := cm.currentProxy.GetProxyConfig().handlerMgr
		if handler == nil {
			hwlog.RunLog.Errorf("message handler is not initialized")
			continue
		}
		go func() {
			cm.sendMsg(handler.HandleMsg(msg, peerInfo))
		}()
	}
}

func (cm *wsConnectMgr) postProcess() {
	for _, f := range cm.currentProxy.GetDisconnectCallbacks() {
		f(cm.peerInfo)
	}
	if bandwidthLimiter := cm.currentProxy.GetBandwidthLimiter(); bandwidthLimiter != nil {
		bandwidthLimiter.UnregisterConn(cm.getPeerId())
	}
	if cm.rpsLimiter != nil {
		cm.rpsLimiter = nil
	}
}

func (cm *wsConnectMgr) readMsg() ([]byte, error) {
	localConn := cm.conn
	if localConn == nil {
		return nil, errors.New("connection is nil")
	}
	messageType, reader, err := localConn.NextReader()
	if err != nil {
		return nil, fmt.Errorf("read message header error: %v", err)
	}
	if messageType != websocket.TextMessage {
		hwlog.RunLog.Errorf("[%s] received not support message type: %v", cm.getPeerId(), messageType)
		return nil, nil
	}
	msg, err := io.ReadAll(io.LimitReader(reader, defaultReadSizeLimit))
	if err != nil {
		return nil, fmt.Errorf("read message body error: %v", err)
	}
	return msg, nil
}

func (cm *wsConnectMgr) sendMsg(data []byte) {
	if len(data) == 0 {
		return
	}
	if err := cm.send(websocket.TextMessage, data); err != nil {
		hwlog.RunLog.Errorf("send resp msg failed: %v", err)
		return
	}
	hwlog.RunLog.Infof("send resp msg success")
}

func (cm *wsConnectMgr) limiterCheck(dataLen int) error {
	if cm.rpsLimiter != nil && !cm.rpsLimiter.Allow() {
		return fmt.Errorf("message process is denied by rps limiter")
	}
	bandwidthLimiter := cm.currentProxy.GetBandwidthLimiter()
	if bandwidthLimiter != nil && !bandwidthLimiter.Allow(cm.getPeerId(), dataLen) {
		return fmt.Errorf("message process is denied by bandwidth limiter")
	}
	return nil
}

func (cm *wsConnectMgr) heartbeat() {
	cm.waitHandleGrp.Add(1)
	for {
		select {
		case <-cm.ctx.Done():
			hwlog.RunLog.Warnf("connect (%s) heartbeat Stop", cm.getPeerId())
			cm.waitHandleGrp.Done()
			return
		default:
		}
		cm.sendPingMsg()
		time.Sleep(defaultHeartbeatInterval)
		if time.Now().Sub(cm.lastAlive) > defaultHeartbeatTimeout {
			hwlog.RunLog.Errorf("[%s] heartbeat timeout", cm.getPeerId())
			cm.waitHandleGrp.Done()
			if err := cm.stop(); err != nil {
				hwlog.RunLog.Errorf("stop connection manager error: %v", err)
			}
			return
		}
	}
}

func (cm *wsConnectMgr) sendPingMsg() {
	if err := cm.send(websocket.PingMessage, []byte(pingMsg)); err != nil {
		hwlog.RunLog.Errorf("send ping frame to peer error: %v", err)
	}
}

func (cm *wsConnectMgr) send(msgType int, data []byte) error {
	if !cm.isConnected() {
		return fmt.Errorf("[%s] websocket not connect, please connect first", cm.getPeerId())
	}
	cm.sendLock.Lock()
	defer cm.sendLock.Unlock()
	localConn := cm.conn
	if localConn == nil {
		return errors.New("websocket connection is nil")
	}
	return localConn.WriteMessage(msgType, data)
}

// according to rfc6455#section-5.5.2, upon receiving a Ping frame, an endpoint MUST send a Pong frame in response.
// A Pong frame sent in response to a Ping frame must have identical "Application data" as found
// in the message body of the Ping frame being replied to.
func (cm *wsConnectMgr) pingHandle(appData string) error {
	select {
	case <-cm.ctx.Done():
		return errors.New("websocket is already closed")
	default:
	}
	cm.lastAlive = time.Now()
	if err := cm.send(websocket.PongMessage, []byte(appData)); err != nil {
		hwlog.RunLog.Errorf("resp pong frame error: %v", err)
	}
	return nil
}

// for compatible with old version software, we refresh alive time upon any Pong frame,
// in future, we should refresh alive time ONLY when appData == pingMsg (message content sent by MEF ping frame)
func (cm *wsConnectMgr) pongHandle(appData string) error {
	select {
	case <-cm.ctx.Done():
		return errors.New("websocket is already closed")
	default:
	}
	cm.lastAlive = time.Now()
	return nil
}

// according to rfc6455#section-5.5.1, if an endpoint receives a Close frame and did not previously send a
// Close frame, the endpoint MUST send a Close frame in response
func (cm *wsConnectMgr) closeHandle(code int, text string) error {
	if err := cm.send(websocket.CloseMessage, websocket.FormatCloseMessage(code, text)); err != nil {
		hwlog.RunLog.Errorf("resp close frame error: %v", err)
	}
	return nil
}

func (cm *wsConnectMgr) stop() error {
	if cm.conn == nil && cm.connFlag == false {
		return nil
	}
	cm.connLock.Lock()
	defer cm.connLock.Unlock()
	hwlog.RunLog.Errorf("[%s] Stop websocket connection", cm.getPeerId())
	cm.cancel()
	cm.waitHandleGrp.Wait()
	if cm.conn == nil {
		return nil
	}
	err := cm.conn.Close()
	cm.conn = nil
	cm.connFlag = false
	return err
}
