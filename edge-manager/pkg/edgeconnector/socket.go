// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeconnector the websocket module config
package edgeconnector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"edge-manager/pkg/database"
	"edge-manager/pkg/edgeinstaller"
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
	server    *WebSocketServer
	writeLock sync.RWMutex
	ctx       context.Context
	enable    bool
	engine    *gin.Engine
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
		case common.Get:
			s.sendInnerGet(message)
		default:
			hwlog.RunLog.Error("invalid operation")
			continue
		}

	}
}

func (s *Socket) sendInnerGet(message *model.Message) {
	tokenReq, ok := message.GetContent().(edgeinstaller.TokenReq)
	if !ok {
		hwlog.RunLog.Error("convert to tokenReq failed")
		return
	}

	if err := s.sendToInstaller(tokenReq.NodeId, message); err != nil {
		hwlog.RunLog.Errorf("construct content or send to edge-installer failed, error: %v", err)
		return
	}

	return
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
	hwlog.RunLog.Info("edge-connector receive message from edge-installer success for upgrading")
	upgradeInfo, ok := message.GetContent().(util.DealSfwContent)
	if !ok {
		hwlog.RunLog.Error("convert to upgradeInfo failed")
		return
	}

	if err := checkUpgradeInfo(upgradeInfo); err != nil {
		hwlog.RunLog.Errorf("check software manager info failed, error: %v", err)
		return
	}
	defer common.ClearStringMemory(upgradeInfo.Password)

	if err := s.sendToInstaller(upgradeInfo.NodeId, message); err != nil {
		hwlog.RunLog.Errorf("construct content or send to edge-installer failed, error: %v", err)
		return
	}

	hwlog.RunLog.Info("edge-connector send upgrade message to edge-installer success for upgrading")
	hwlog.RunLog.Info(" ----------deal upgrade request from edge-installer end-------")
	return
}

func (s *Socket) sendInnerDownload(message *model.Message) {
	hwlog.RunLog.Info("edge-connector receive message from edge-installer success for downloading")
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

	hwlog.RunLog.Info("edge-connector send message to edge-installer success for upgrading")
	hwlog.RunLog.Info(" ----------deal download request from edge-installer end-------")
	return
}

func (s *Socket) sendInnerUpdate(message *model.Message) {
	hwlog.RunLog.Info(" ----------deal update request from restful begin-------")
	if !(message.GetSource() == common.RestfulServiceName) {
		hwlog.RunLog.Error("invalid source when updating")
		return
	}

	if err := s.respRestful(message); err != nil {
		hwlog.RunLog.Error("send response for updating to restful module failed")
		return
	}
	hwlog.RunLog.Info("edge-connector send SUCCESS to restful for updating success")

	var updateInfo UpdateInfo
	if err := common.ParamConvert(message.GetContent(), &updateInfo); err != nil {
		hwlog.RunLog.Error("convert to updateInfo failed")
		return
	}
	defer common.ClearStringMemory(updateInfo.Password)
	updateInfoToInstaller := &UpdateInfoToInstaller{
		Username: updateInfo.Username,
		Password: []byte(updateInfo.Password),
	}
	message.FillContent(updateInfoToInstaller)

	nodeIds, err := getUniqueNums()
	if err != nil {
		hwlog.RunLog.Error("get unique nums failed")
		return
	}

	for _, nodeId := range nodeIds {
		go s.sendToEdge(nodeId, message)
	}

	hwlog.RunLog.Info("edge-connector send message to edge-installer success for updating")
	hwlog.RunLog.Info(" ----------deal update request from edge-installer end-------")
	return
}

func (s *Socket) sendToEdge(nodeId string, message *model.Message) {
	if err := s.sendToInstaller(nodeId, message); err != nil {
		hwlog.RunLog.Errorf("construct content or send to edge-installer failed, error: %v", err)
		return
	}
	return
}

func (s *Socket) respRestful(message *model.Message) error {
	respToRestful, respContent := s.constructContent(message)
	respToRestful.FillContent(respContent)
	if err := modulemanager.SendMessage(respToRestful); err != nil {
		hwlog.RunLog.Errorf("%s send response to restful failed", common.EdgeConnectorName)
		return err
	}

	hwlog.RunLog.Info("edge-connector update user info in table conn_infos success")
	return nil
}

func (s *Socket) constructContent(message *model.Message) (*model.Message, common.RespMsg) {
	var req UpdateInfo
	if err := common.ParamConvert(message.GetContent(), &req); err != nil {
		return nil, common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	respMsg := UpdateUserConnInfo(req)
	respToRestful, err := message.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("%s new response failed", s.Name())
		return nil, common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	return respToRestful, respMsg
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
		s.server.allClients.Range(func(k, v interface{}) bool {
			nodeId, ok := k.(string)
			if !ok {
				hwlog.RunLog.Error("convert key from clients map to string failed")
				return false
			}

			conn, ok := v.(*websocket.Conn)
			if !ok {
				hwlog.RunLog.Error("convert value from clients map to string failed")
				return false
			}
			go s.dealConnMap(nodeId, conn)
			return true
		})
		<-s.server.loopChan
		s.cleanChannel()
	}
}

func (s *Socket) cleanChannel() {
	for {
		select {
		case _, ok := <-s.server.loopChan:
			if !ok {
				hwlog.RunLog.Error("loopChan has been closed")
				return
			}
		default:
			return
		}
	}
}

func (s *Socket) dealFailedNode(failedNode string) {
	s.writeLock.Lock()
	s.server.isConnMap.Store(failedNode, false)
	s.writeLock.Unlock()
	s.server.loopChan <- struct{}{}
	return
}

func (s *Socket) dealConnMap(nodeId string, conn *websocket.Conn) {
	isConn, ok := s.server.isConnMap.Load(nodeId)
	if !ok {
		hwlog.RunLog.Error("load connection state from isConn map failed ")
		return
	}

	if !isConn.(bool) {
		hwlog.RunLog.Errorf("edge-installer [%s] is disconnected", nodeId)
		return
	}

	if !s.dealConn(nodeId, conn) { // todo 多节点会阻塞
		s.dealFailedNode(nodeId)
	}
	return
}

func (s *Socket) dealConn(nodeId string, conn *websocket.Conn) bool {
	var message *model.Message
	var err error

	for {
		select {
		case _, ok := <-s.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("has listened stop signal")
			return true
		default:
		}

		if message, err = s.server.Receive(conn); err != nil {
			hwlog.RunLog.Errorf("receive message from edge-installer failed, error: %v", err)
			return false
		}

		hwlog.RunLog.Infof("edge-connector receive message from edge-installer [%s]", nodeId)
		go s.dealMessageFromInstaller(message, nodeId)
	}
}

func (s *Socket) dealMessageFromInstaller(message *model.Message, nodeId string) {
	if message == nil {
		hwlog.RunLog.Error("receive message is nil")
		return
	}

	hwlog.RunLog.Infof("edge-connector receive message from edge-installer, Operation: [%+v]; Resource: [%+v]",
		message.GetOption(), message.GetResource())

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
	case common.Get:
		s.dealExternalGet(nodeId, message)
	default:
		hwlog.RunLog.Error("invalid option")
		return
	}
}

func (s *Socket) dealExternalGet(nodeId string, message *model.Message) {
	tokenReq := &edgeinstaller.TokenReq{
		NodeId: nodeId,
	}
	content, err := json.Marshal(tokenReq)
	if err != nil {
		hwlog.RunLog.Errorf("marshal content failed, error: %v", err)
		return
	}

	message.SetRouter(s.Name(), common.EdgeInstallerName, common.Get, common.Token)
	s.sendToModule(string(content), message)
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
