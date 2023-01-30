// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package websocketmgr

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"huawei.com/mindx/common/hwlog"
)

type wsConnectMgr struct {
	conn          *websocket.Conn
	name          string
	handler       HandleMsgIntf
	lastAlive     time.Time
	connFlag      bool
	waitHandleGrp sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	sendLock      sync.Mutex
}

func (cm *wsConnectMgr) start(conn *websocket.Conn, name string, handler HandleMsgIntf) {
	cm.conn = conn
	cm.name = name
	cm.handler = handler
	cm.ctx, cm.cancel = context.WithCancel(context.Background())
	cm.connFlag = true
	cm.lastAlive = time.Now()
	cm.conn.SetCloseHandler(cm.closeHandle)
	cm.conn.SetPongHandler(cm.pongHandle)
	cm.conn.SetPingHandler(cm.pingHandle)
	cm.conn.SetReadLimit(defaultReadLimit)
	cm.startLoop()
}

func (cm *wsConnectMgr) getConnName() string {
	return cm.name
}

func (cm *wsConnectMgr) isConnected() bool {
	return cm.connFlag && cm.conn != nil
}

func (cm *wsConnectMgr) startLoop() {
	go func() {
		cm.waitHandleGrp.Add(1)
		defer cm.waitHandleGrp.Done()
		cm.heartbeat()
	}()

	go func() {
		cm.waitHandleGrp.Add(1)
		defer cm.waitHandleGrp.Done()
		cm.receive()
	}()
}

func (cm *wsConnectMgr) receive() {
	for {
		select {
		case <-cm.ctx.Done():
			return
		default:
		}
		messageType, reader, err := cm.conn.NextReader()
		if err != nil {
			hwlog.RunLog.Errorf("[%s] websocket NextReader error: %v", cm.name, err)
			err := cm.stop()
			if err != nil {
				return
			}
		}
		if messageType != websocket.TextMessage {
			hwlog.RunLog.Errorf("[%s] received not support message type: %v", cm.name, messageType)
			continue
		}
		msg, err := io.ReadAll(io.LimitReader(reader, defaultReadLimit))
		if err != nil {
			hwlog.RunLog.Errorf("[%s] read msg from reader error: %v", cm.name, err)
			continue
		}
		go cm.handler.handleMsg(msg)
	}
}

func (cm *wsConnectMgr) heartbeat() {
	for {
		select {
		case <-cm.ctx.Done():
			return
		default:
		}
		cm.sendPingMsg()
		time.Sleep(defaultHeartbeatDuration)
		if time.Now().Sub(cm.lastAlive) > defaultHeartbeatTimeout {
			hwlog.RunLog.Errorf("[%s] heartbeat timeout", cm.name)
			err := cm.stop()
			if err != nil {
				return
			}
		}
	}
}

func (cm *wsConnectMgr) pingHandle(appData string) error {
	cm.lastAlive = time.Now()
	hwlog.RunLog.Debugf("[%s] received ping message: %v", cm.name, appData)
	return nil
}

func (cm *wsConnectMgr) sendPingMsg() {
	err := cm.send(wsMessage{MsgType: websocket.PongMessage, Value: []byte(cm.name + ":send ping to pong")})
	if err != nil {
		return
	}
}

func (cm *wsConnectMgr) send(msg wsMessage) error {
	cm.sendLock.Lock()
	defer cm.sendLock.Unlock()
	if !cm.isConnected() {
		return fmt.Errorf("[%s] websocket not connect, please connect first", cm.name)
	}

	err := cm.conn.WriteMessage(msg.MsgType, msg.Value)
	if err != nil {
		return err
	}
	return nil
}

func (cm *wsConnectMgr) pongHandle(appData string) error {
	cm.lastAlive = time.Now()
	hwlog.RunLog.Debugf("[%s] received pong message: %v", cm.name, appData)
	return nil
}

func (cm *wsConnectMgr) closeHandle(code int, text string) error {
	hwlog.RunLog.Errorf("[%s] websocket connection closed, code:%d, text:%v", cm.name, code, text)
	err := cm.stop()
	if err != nil {
		return err
	}
	return nil
}

func (cm *wsConnectMgr) stop() error {
	hwlog.RunLog.Errorf("[%s] stop websocket connection", cm.name)
	cm.connFlag = false
	cm.cancel()
	if cm.conn != nil {
		err := cm.conn.Close()
		if err != nil {
			return err
		}
	}
	cm.conn = nil
	cm.waitHandleGrp.Wait()
	return nil
}
