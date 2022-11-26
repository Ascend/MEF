// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeconnector the websocket module config
package edgeconnector

import (
	"context"

	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"

	"edge-manager/pkg/database"
	"edge-manager/pkg/util"

	"github.com/gorilla/websocket"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
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
	s.server = newWebsocketServer()

	s.start()
}

func (s *Socket) start() {
	go s.server.startWebsocketServer()
	go s.receiveAndSendInnerMsg()
	go s.receiveAndSendExternalMsg()
}

func (s *Socket) receiveAndSendInnerMsg() {
	hwlog.RunLog.Info("start receive and send inner message")
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

		message, err := modulemanager.ReceiveMessage(s.Name())
		if err != nil {
			hwlog.RunLog.Errorf("receive message from channel failed, error: %v", err)
			continue
		}
		if !util.CheckInnerMsg(message) {
			hwlog.RunLog.Error("message receive from module is invalid")
			continue
		}
		if message.GetDestination() != common.EdgeConnectorName {
			hwlog.RunLog.Errorf("message is not sent to edge-connector, destination is: %s", message.GetDestination())
			return
		}

		switch message.GetOption() {
		case common.Issue:
			s.sendInnerIssue(message)
		case common.Upgrade:
			s.sendInnerUpgrade(message)
		case common.Update:
			s.sendInnerUpdate(message)
		default:
			hwlog.RunLog.Error("invalid option")
			return
		}
		return
	}
}

// IssueInfo struct for issuing service cert
type IssueInfo struct {
	NodeId      string
	ServiceCert []byte
}

func (s *Socket) sendInnerIssue(message *model.Message) {
	issueInfo, ok := message.GetContent().(IssueInfo)
	if !ok {
		hwlog.RunLog.Error("convert to issueInfo failed")
		return
	}

	if !s.server.IsClientConnected(issueInfo.NodeId) {
		hwlog.RunLog.Errorf("edge-installer %s is disconnected", issueInfo.NodeId)
		return
	}
	s.sendToInstaller(issueInfo.NodeId, message)

	return
}

// UpgradeInfo struct for upgrading software
type UpgradeInfo struct {
	NodeId          []string
	SoftwareName    string
	SoftwareVersion string
	baseInfo
}

func (s *Socket) sendInnerUpgrade(message *model.Message) {
	upgradeInfo, ok := message.GetContent().(UpgradeInfo)
	if !ok {
		hwlog.RunLog.Error("convert to upgradeInfo failed")
		return
	}

	if err := upgradeInfo.checkBaseInfo(); err != nil {
		return
	}

	for _, nodeId := range upgradeInfo.NodeId {
		if !s.server.IsClientConnected(nodeId) {
			hwlog.RunLog.Errorf("edge-installer %s is disconnected", nodeId)
			return
		}
		s.sendToInstaller(nodeId, message)
	}

	return
}

// UpdateInfo struct for updating username and password
type UpdateInfo struct {
	NodeId   []string
	Username string
	Password []byte
}

func (s *Socket) sendInnerUpdate(message *model.Message) {
	updateInfo, ok := message.GetContent().(UpdateInfo)
	defer common.ClearSliceByteMemory(updateInfo.Password)
	if !ok {
		hwlog.RunLog.Error("convert to upgradeInfo failed")
		return
	}

	for _, nodeId := range updateInfo.NodeId {
		if !s.server.IsClientConnected(nodeId) {
			hwlog.RunLog.Errorf("edge-installer %s is disconnected", nodeId)
			return
		}
		s.sendToInstaller(nodeId, message)
	}

	return
}

func (s *Socket) sendToInstaller(nodeId string, message *model.Message) {
	sendMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("new message failed, error: %v", err)
		return
	}
	sendMsg.SetRouter(common.EdgeConnectorName, common.EdgeInstallerName, message.GetOption(), message.GetResource())
	sendMsg.FillContent(message.GetContent())
	sendMsg.SetIsSync(false)
	if err = s.server.Send(sendMsg, nodeId); err != nil {
		s.server.CloseConnection(nodeId)
		return
	}

	return
}

func (s *Socket) receiveAndSendExternalMsg() {
	hwlog.RunLog.Info("start receive and send external message")
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

		for _, conn := range s.server.allClients {
			go s.dealConnMap(conn)
		}
	}
}

func (s *Socket) dealConnMap(conn *websocket.Conn) {
	var message *model.Message
	var err error

	if message, err = s.server.Receive(conn); err != nil {
		return
	}

	if message == nil {
		hwlog.RunLog.Error("receive message is nil")
		return
	}

	if !util.CheckInnerMsg(message) {
		hwlog.RunLog.Error("message receive from edge-installer is invalid")
		return
	}

	if message.GetDestination() != common.EdgeConnectorName {
		hwlog.RunLog.Errorf("message is not sent to edge-connector, destination is: %s", message.GetDestination())
		return
	}

	switch message.GetOption() {
	case common.Issue:
		s.dealExternalIssue(message)
	case common.Upgrade:
		s.dealExternalUpgrade(message)
	default:
		hwlog.RunLog.Error("invalid option")
		return
	}
	return
}

// IssueResp deals issue service cert response
type IssueResp struct {
	NodeId []string
	Result string
	Reason string
}

func (s *Socket) dealExternalIssue(message *model.Message) {
	issueResp, ok := message.GetContent().(IssueResp)
	if !ok {
		hwlog.RunLog.Error("convert to updateResp failed")
		return
	}

	s.sendToModule(issueResp, message)
	return
}

// UpgradeResp deals upgrade software response
type UpgradeResp struct {
	NodeId []string
	Result string
	Reason string
}

func (s *Socket) dealExternalUpgrade(message *model.Message) { // 后续需传回restful
	updateResp, ok := message.GetContent().(UpgradeResp)
	if !ok {
		hwlog.RunLog.Error("convert to updateResp failed")
		return
	}

	if updateResp.Result == "fail" {
		hwlog.RunLog.Errorf("edge-installer %s upgrade software failed, reason: %s",
			updateResp.NodeId, updateResp.Reason)
		return
	}
	hwlog.RunLog.Infof("edge-installer %s upgrade software success", updateResp.NodeId)
	return
}

func (s *Socket) sendToModule(input interface{}, message *model.Message) {
	destination := getDestination(message)

	sendMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("new message failed, error: %v", err)
		return
	}
	sendMsg.SetRouter(common.EdgeConnectorName, destination, message.GetOption(), message.GetResource())
	sendMsg.FillContent(input)
	sendMsg.SetIsSync(false)
	if err = modulemanager.SendMessage(sendMsg); err != nil {
		hwlog.RunLog.Errorf("send message to module failed, error: %v", err)
		return
	}

	return
}

func getDestination(message *model.Message) string {
	var destination = ""
	switch message.GetOption() {
	case common.Issue:
		destination = common.CertManagerName
	case common.Upgrade:
		destination = common.RestfulServiceName
	default:
		hwlog.RunLog.Error("invalid option")
		return ""
	}
	return destination
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
