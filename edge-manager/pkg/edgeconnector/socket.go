// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeconnector the websocket module config
package edgeconnector

import (
	"context"
	"edge-manager/module_manager"
	"edge-manager/module_manager/model"
	"edge-manager/pkg/common"
	"edge-manager/pkg/database"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"huawei.com/mindx/common/hwlog"
)

// Socket wraps the struct WebSocketServer
type Socket struct {
	server *WebSocketServer
	ctx    context.Context
	enable bool
}

// Name returns the name of websocket connection module
func (s *Socket) Name() string {
	return common.EdgeConnectorName
}

// Start initializes the websocket server
func (s *Socket) Start() {
	if err := database.CreateTableIfNotExists(ConnInfo{}); err != nil {
		hwlog.RunLog.Error("table conn_infos create failed")
		return
	}

	if err := s.initSocket(); err != nil {
		hwlog.RunLog.Errorf("start websocket server failed, error: %v", err)
		return
	}
	go s.server.startWebsocketServer()
}

func (s *Socket) initSocket() error {
	var err error
	s.server, err = newWebsocketServer()
	if err != nil {
		return err
	}
	return nil
}

// UpgradeInfo struct for updating software
type UpgradeInfo struct {
	NodeId []string
	baseInfo
}

func (s *Socket) receiveFromModule() {
	for {
		select {
		case _, ok := <-s.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("has listened stop signal")
			return
		default:
		}
		msg, err := module_manager.ReceiveMessage(s.Name())
		if err != nil {
			hwlog.RunLog.Errorf("receive message from channel failed, error: %v", err)
			continue
		}
		if !common.CheckInnerMsg(msg) {
			hwlog.RunLog.Error("message receive from module is invalid")
			continue
		}
		upgradeInfo, ok := msg.GetContent().(UpgradeInfo)
		if !ok {
			hwlog.RunLog.Error("convert to UpgradeInfo failed")
			continue
		}

		if err = upgradeInfo.checkBaseInfo(); err != nil {
			continue
		}
		for _, nodeId := range upgradeInfo.NodeId {
			if !s.server.IsClientConnected(nodeId) {
				hwlog.RunLog.Errorf("edge-installer %s is disconnected", nodeId)
				continue
			}
			s.sendToInstaller(nodeId, msg)
			go s.receiveFromInstaller(nodeId)
		}
	}
}

func (s *Socket) sendToInstaller(nodeId string, msg *model.Message) {
	if msg.GetDestination() != common.EdgeInstallerName {
		hwlog.RunLog.Errorf("message is not sent to edge-installer, destination is: %s", msg.GetDestination())
		return
	}
	conn := s.server.allClients[nodeId]
	if err := conn.WriteMessage(websocket.TextMessage, msg.GetContent().([]byte)); err != nil {
		hwlog.RunLog.Errorf("write message failed, error: %v", err)
		s.server.CloseConnection(nodeId)
		return
	}
	return
}

// DealSfwResp deals software response
type DealSfwResp struct {
	Operation string `json:"operation"`
	Result    string `json:"result"`
	Reason    string `json:"reason"`
}

func (s *Socket) receiveFromInstaller(nodeId string) {
	conn := s.server.allClients[nodeId]
	if err := conn.SetReadDeadline(time.Now().Add(ReadInstallerDeadline)); err != nil {
		hwlog.RunLog.Error("read message time out")
		return
	}
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			hwlog.RunLog.Errorf("websocket server read error: %v", err)
			continue
		}
		resp := &DealSfwResp{}
		if err = json.Unmarshal(data, &resp); err != nil {
			hwlog.RunLog.Errorf("parse message failed, error: %v", err)
			continue
		}

		if resp.Result == "fail" {
			hwlog.RunLog.Errorf("edge-installer %s %s software failed, reason: %s",
				nodeId, resp.Operation, resp.Reason)
		}
		hwlog.RunLog.Infof("edge-installer %s %s software success", nodeId, resp.Operation)
		return
	}
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

// NewSocket new Socket
func NewSocket(enable bool) *Socket {
	socket := &Socket{
		enable: enable,
		ctx:    context.Background(),
	}
	return socket
}
