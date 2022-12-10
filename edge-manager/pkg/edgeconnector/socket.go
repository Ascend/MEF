// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeconnector the websocket module config
package edgeconnector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"edge-manager/pkg/database"
	"edge-manager/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

// Socket wraps the struct WebSocketServer
type Socket struct {
	server *WebSocketServer
	ctx    context.Context
	enable bool
	engine *gin.Engine
}

// Name returns the name of websocket connection module
func (s *Socket) Name() string {
	return common.EdgeConnectorName
}

// Start initializes the websocket server
func (s *Socket) Start() {
	InitConfigure()
	s.server = newWebsocketServer(s.engine)
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
		if message.GetDestination() != s.Name() {
			hwlog.RunLog.Errorf("message is not sent to edge-connector, destination is: %s", message.GetDestination())
			continue
		}

		switch message.GetOption() {
		case common.Issue:
			s.sendInnerIssue(message)
		case common.Upgrade:
			s.sendInnerUpgrade(message)
		case common.Download:
			s.sendInnerDownload(message)
		case common.Update:
			s.sendInnerUpdate(message)
		default:
			hwlog.RunLog.Error("invalid operation")
			continue
		}

	}
}

func (s *Socket) sendInnerIssue(message *model.Message) {
	issueInfo, ok := message.GetContent().(IssueInfo)
	if !ok {
		hwlog.RunLog.Error("convert to issueInfo failed")
		return
	}

	if err := s.sendToInstaller(issueInfo.NodeId, message); err != nil {
		hwlog.RunLog.Errorf("construct content or send to edge-installer failed, error: %v", err)
		return
	}

	return
}

func (s *Socket) sendInnerUpgrade(message *model.Message) {
	upgradeInfo, ok := message.GetContent().(util.DealSfwContent)
	if !ok {
		hwlog.RunLog.Error("convert to upgradeInfo failed")
		return
	}

	if err := checkUpgradeInfo(upgradeInfo); err != nil {
		hwlog.RunLog.Errorf("check software manager info failed, error: %v", err)
		return
	}

	if err := s.sendToInstaller(upgradeInfo.NodeId, message); err != nil {
		hwlog.RunLog.Errorf("construct content or send to edge-installer failed, error: %v", err)
		return
	}

	return
}

func (s *Socket) sendInnerDownload(message *model.Message) {
	hwlog.RunLog.Info("edge-connector receive message from edge-installer success")
	downloadSfw, ok := message.GetContent().(util.DealSfwContent)
	defer common.ClearStringMemory(downloadSfw.Password)
	if !ok {
		hwlog.RunLog.Error("convert to downloadSfw failed")
		return
	}

	if err := s.sendToInstaller(downloadSfw.NodeId, message); err != nil {
		hwlog.RunLog.Errorf("construct content or send to edge-installer failed, error: %v", err)
		return
	}

	hwlog.RunLog.Info("edge-connector send message to edge-installer success")
	hwlog.RunLog.Info(" ----------deal download request from edge-installer end-------")
	return
}

func (s *Socket) sendInnerUpdate(message *model.Message) {
	updateInfo, ok := message.GetContent().(UpdateInfo)
	defer common.ClearSliceByteMemory(updateInfo.Password)
	if !ok {
		hwlog.RunLog.Error("convert to upgradeInfo failed")
		return
	}

	for _, nodeId := range updateInfo.NodeId {
		if err := s.sendToInstaller(nodeId, message); err != nil {
			hwlog.RunLog.Errorf("construct content or send to edge-installer failed, error: %v", err)
			return
		}
	}

	return
}

func (s *Socket) sendToInstaller(nodeId string, message *model.Message) error {
	if !s.server.IsClientConnected(nodeId) {
		hwlog.RunLog.Errorf("edge-installer %s is disconnected", nodeId)
		return fmt.Errorf("edge-installer %s is disconnected", nodeId)
	}

	data, err := json.Marshal(message.GetContent())
	if err != nil {
		hwlog.RunLog.Errorf("marshal content failed, error: %v", err)
		return err
	}

	sendMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("new message failed, error: %v", err)
		return err
	}
	sendMsg.SetRouter(common.EdgeConnectorName, common.EdgeInstallerName, message.GetOption(), message.GetResource())
	sendMsg.FillContent(string(data))
	sendMsg.SetIsSync(false)
	if err = s.server.Send(sendMsg, nodeId); err != nil {
		s.server.CloseConnection(nodeId)
		hwlog.RunLog.Errorf("send message to edge-installer failed, error: %v", err)
		return err
	}

	return nil
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

		for nodeId, conn := range s.server.allClients {
			if !s.server.isConnMap[nodeId] {
				continue
			} else {
				hwlog.RunLog.Infof("edge-connector receive message from edge-installer [%s]", nodeId)
				s.dealConnMap(conn)
			}
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

	hwlog.RunLog.Infof("edge-connector receive message from edge-installer, Source: [%+v]; Destination: [%+v]",
		message.GetSource(), message.GetDestination())

	if !util.CheckInnerMsg(message) {
		hwlog.RunLog.Error("message receive from edge-installer is invalid")
		return
	}

	switch message.GetOption() {
	case common.Issue:
		s.dealExternalIssue(message)
	case common.Upgrade:
		s.dealExternalUpgrade(message)
	case common.Download:
		s.dealExternalDownload(message)
	default:
		hwlog.RunLog.Error("invalid option")
		return
	}
	return
}

func (s *Socket) dealExternalIssue(message *model.Message) {
	content, ok := message.GetContent().(string)
	if !ok {
		hwlog.RunLog.Error("convert to content error")
		return
	}

	var issueResp IssueResp
	if err := json.Unmarshal([]byte(content), &issueResp); err != nil {
		hwlog.RunLog.Error("parse to data failed")
		return
	}

	s.sendToModule(issueResp, message)
	return
}

func (s *Socket) dealExternalDownload(message *model.Message) {
	content, ok := message.GetContent().(string)
	if !ok {
		hwlog.RunLog.Error("convert to content failed")
		return
	}

	var data interface{}
	switch message.GetResource() {
	case common.Software:
		hwlog.RunLog.Info("----------deal download request from edge-installer begin-------")
		parseContent := isDownloadReq(content)
		if parseContent == nil {
			return
		}
		data = *parseContent
	case common.SoftwareResp:
		hwlog.RunLog.Info("----------deal download response from edge-installer begin-------")
		parseContent := isDownloadResp(content)
		if parseContent == nil {
			return
		}
		data = *parseContent
	default:
		hwlog.RunLog.Error("invalid resource")
	}

	s.sendToModule(data, message)
	return
}

func (s *Socket) dealExternalUpgrade(message *model.Message) { // todo 后续需传回restful
	content, ok := message.GetContent().(string)
	if !ok {
		hwlog.RunLog.Error("convert to content error")
		return
	}

	var respFromInstaller RespFromInstaller
	if err := json.Unmarshal([]byte(content), &respFromInstaller); err != nil {
		hwlog.RunLog.Error("parse to data failed")
		return
	}

	if respFromInstaller.Result == common.FailResult {
		hwlog.RunLog.Errorf("edge-installer %s upgrade software failed, reason: %s",
			respFromInstaller.NodeId, respFromInstaller.Reason)
		return
	}

	hwlog.RunLog.Infof("--------edge-installer %s upgrade software success--------", respFromInstaller.NodeId)
	return
}

func (s *Socket) sendToModule(input interface{}, message *model.Message) {
	destination := getDestination(message)

	sendMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("new message failed, error: %v", err)
		return
	}
	sendMsg.SetRouter(s.Name(), destination, message.GetOption(), message.GetResource())
	sendMsg.FillContent(input)
	sendMsg.SetIsSync(false)
	if err = modulemanager.SendMessage(sendMsg); err != nil {
		hwlog.RunLog.Errorf("send message to module [%s] failed, error: %v", destination, err)
		return
	}

	hwlog.RunLog.Infof("send [%s] message to module [%s] success", message.GetOption(), destination)
	return
}

// Enable indicates whether this module is enabled
func (s *Socket) Enable() bool {
	if s.enable {
		if err := initConnInfoTable(); err != nil {
			hwlog.RunLog.Errorf("module (%s) init database table failed, cannot enable", common.EdgeConnectorName)
			return !s.enable
		}
	}

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
		engine: initGin(),
	}
	return socket
}

func initConnInfoTable() error {
	if err := database.CreateTableIfNotExists(ConnInfo{}); err != nil {
		return errors.New("create database table conn_infos failed")
	}
	return nil
}
