// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeconnector the websocket module config
package edgeconnector

import (
	"huawei.com/mindx/common/hwlog"
)

// Socket wraps the struct WebSocketServer
type Socket struct {
	server *WebSocketServer
	enable bool
}

// Name returns the name of websocket connection module
func (s *Socket) Name() string {
	return "center-edge connector"
}

// Start initializes the websocket server
func (s *Socket) Start() {
	if err := s.initSocket(); err != nil {
		hwlog.RunLog.Error("start websocket server failed, error: ", err)
	}
	go s.server.StartWebsocketServer()
}

func (s *Socket) initSocket() error {
	var err error
	s.server, err = NewWebsocketServer()
	if err != nil {
		return err
	}
	return nil
}

// SendToEdge sends message to edge-installer
func (s *Socket) SendToEdge(nodeID string) {
	hwlog.RunLog.Info("start sending message to edge-installer")
	if !s.server.IsClientConnected(nodeID) {
		hwlog.RunLog.Errorf("edge-installer %v is disconnected", nodeID)
		return
	}
	s.dispatch(nodeID)
}

func (s *Socket) dispatch(nodeID string) {

}

// Enable indicates whether this module is enabled
func (s *Socket) Enable() bool {
	return s.enable
}

// Register registers websocket connection module
func Register() *Socket {
	socket := enaSocket(true)
	return socket
}

func enaSocket(enable bool) *Socket {
	return &Socket{
		enable: enable,
	}
}
