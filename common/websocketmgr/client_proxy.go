// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package websocketmgr for websocket manager
package websocketmgr

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
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

// GetName get websocket client name
func (wcp *WsClientProxy) GetName() string {
	return wcp.ProxyCfg.name
}

// Start websocket client start
func (wcp *WsClientProxy) Start() error {
	dialer := &websocket.Dialer{
		TLSClientConfig:  wcp.ProxyCfg.tlsConfig,
		HandshakeTimeout: handshakeTimeout,
		ReadBufferSize:   readBufferSize,
		WriteBufferSize:  writeBufferSize,
	}
	connect, err := wcp.tryConnect(dialer)
	if err != nil {
		return fmt.Errorf("connect the server failed: %v", err)
	}
	wcp.connMgr = &wsConnectMgr{}
	wcp.connMgr.start(connect, wcp.ProxyCfg.name, &wcp.ProxyCfg.handlerMgr)
	return nil
}

// Stop websocket client stop
func (wcp *WsClientProxy) Stop() error {
	wcp.ProxyCfg.cancel()
	err := wcp.connMgr.stop()
	if err != nil {
		return fmt.Errorf("stop websocket connection failed: %v", err)
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
	var retErr error
	for i := 0; i < defaultRetryCount; i++ {
		select {
		case <-wcp.ProxyCfg.ctx.Done():
			return nil, fmt.Errorf("connect has be canceled")
		default:
		}
		connect, _, err := dialer.Dial(wsProtocol+wcp.ProxyCfg.hosts, wcp.ProxyCfg.headers)
		retErr = err
		if err == nil {
			return connect, nil
		}
		time.Sleep(retryTime)
	}
	return nil, fmt.Errorf("connect failed: %v", retErr)
}
