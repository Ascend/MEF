// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package cloudhub module edge-connector init
package cloudhub

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/websocketmgr"
	"huawei.com/mindx/common/x509/certutils"

	"edge-manager/pkg/cloudhub/innerwebsocket"
	"edge-manager/pkg/constants"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
)

type messageHandlerFunc func(*model.Message) (*model.Message, bool, error)

type messageHandler struct {
	HandlerFunc messageHandlerFunc
	// indicates whether the message needs to print log, mainly for avoiding recording to many alarm messages
	NeedLogging bool
}

var lockFlag int64 = nolockingFlag

var messageHandlerMap = make(map[string]messageHandler)

func getMsgHandler(msg *model.Message) (messageHandler, bool) {
	handlerKey := msg.GetOption() + msg.GetResource()
	handler, ok := messageHandlerMap[handlerKey]
	return handler, ok
}
func (c *CloudServer) initMsgHandler() {
	messageHandlerMap[common.OptPost+common.ResEdgeCert] = messageHandler{
		HandlerFunc: issueCertForEdge,
		NeedLogging: true,
	}
	messageHandlerMap[common.OptGet+common.ResEdgeConnStatus] = messageHandler{
		HandlerFunc: c.getEdgeConnStatus,
		NeedLogging: true,
	}
	messageHandlerMap[common.OptPost+requests.ReportAlarmRouter] = messageHandler{
		HandlerFunc: innerwebsocket.AlarmReportHandler,
		NeedLogging: false,
	}
	messageHandlerMap[common.Delete+requests.ClearOneNodeAlarmRouter] = messageHandler{
		HandlerFunc: innerwebsocket.AlarmClearHandler,
		NeedLogging: false,
	}
}

// CloudServer wraps the struct WebSocketServer
type CloudServer struct {
	serverIp     string
	wsPort       int
	innerWsPort  int
	authPort     int
	maxClientNum int
	serverProxy  *websocketmgr.WsServerProxy
	writeLock    sync.RWMutex
	ctx          context.Context
	enable       bool
}

var server CloudServer

// NewCloudServer new cloud server
func NewCloudServer(enable bool, wsPort, innerWsPort, authPort, maxClientNum int) *CloudServer {
	server = CloudServer{
		wsPort:       wsPort,
		authPort:     authPort,
		maxClientNum: maxClientNum,
		ctx:          context.Background(),
		enable:       enable,
		innerWsPort:  innerWsPort,
	}
	return &server
}

// Name returns the name of websocket connection module
func (c *CloudServer) Name() string {
	return common.CloudHubName
}

// Enable indicates whether this module is enabled
func (c *CloudServer) Enable() bool {
	if c.enable {
		if err := initNodeTable(); err != nil {
			hwlog.RunLog.Errorf("module (%s) init token database table failed, cannot enable", c.Name())
			return !c.enable
		}
	}
	return c.enable
}

// Start initializes the websocket server
func (c *CloudServer) Start() {
	hwlog.RunLog.Info("----------------cloud hub start----------------")
	checkAndLock()
	var err error
	// set up a websocket server for connecting edge-nodes
	c.serverProxy, err = InitServer()
	if err != nil {
		hwlog.RunLog.Errorf("init websocket server failed: %v", err)
		return
	}
	// set up a websocket connection between edge-manager and alarm manager, for purposes of reporting alarms
	// and inter pods communication
	if err := innerwebsocket.InitInnerWsServer(c.innerWsPort); err != nil {
		hwlog.RunLog.Errorf("init inner websocket server failed: %v", err)
		return
	}
	hwlog.RunLog.Info("init websocket server succeeded")
	c.initMsgHandler()
	for {
		select {
		case _, ok := <-c.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("has listened stop signal")
			return
		default:
		}

		message, err := modulemgr.ReceiveMessage(c.Name())
		if err != nil {
			hwlog.RunLog.Errorf("module [%s] receive message from channel failed, error: %v", c.Name(), err)
			continue
		}

		go c.dispatch(message)
	}
}

func (c *CloudServer) dispatch(message *model.Message) {
	var retMsg = message
	handler, ok := getMsgHandler(message)
	if ok {
		sn := message.GetPeerInfo().Sn
		ip := message.GetPeerInfo().Ip
		if handler.NeedLogging {
			hwlog.RunLog.Infof("%v[%v:%v] %v %v start", time.Now().Format(time.RFC3339Nano),
				ip, sn, message.GetOption(), message.GetResource())
		}
		result, sentToEdge, err := handler.HandlerFunc(message)
		if err != nil {
			if handler.NeedLogging {
				hwlog.RunLog.Errorf("%v [%v:%v] %v %v failed", time.Now().Format(time.RFC3339Nano),
					ip, sn, message.GetOption(), message.GetResource())
			}
			hwlog.RunLog.Errorf("process message [option: %v res: %v]error: %v",
				message.GetOption(), message.GetResource(), err)
			return
		}
		if handler.NeedLogging {
			hwlog.RunLog.Infof("%v [%v:%v] %v %v success", time.Now().Format(time.RFC3339Nano),
				ip, sn, message.GetOption(), message.GetResource())
		}
		if !sentToEdge {
			return
		}
		retMsg = result
	}
	if c.sendToEdge(retMsg) != nil {
		c.response(retMsg, common.FAIL)
	} else {
		c.response(retMsg, common.OK)
	}
}

func (c *CloudServer) response(message *model.Message, content string) {
	if !message.GetIsSync() {
		return
	}

	resp, err := message.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("%s new response failed", c.Name())
		return
	}

	if err = resp.FillContent(content); err != nil {
		hwlog.RunLog.Errorf("fill resp into content failed: %v", err)
		return
	}
	if err = modulemgr.SendMessage(resp); err != nil {
		hwlog.RunLog.Errorf("%s send response failed: %v", c.Name(), err)
	}
}

func (c *CloudServer) sendToEdge(msg *model.Message) error {
	sender := c.serverProxy
	if sender == nil {
		hwlog.RunLog.Errorf("server proxy is not initialized")
		return fmt.Errorf("server proxy is not initialized")
	}

	originalSync := msg.GetIsSync()
	msg.SetIsSync(false)
	defer msg.SetIsSync(originalSync)
	if err := sender.Send(msg.GetNodeId(), msg); err != nil {
		hwlog.RunLog.Errorf("cloud hub send msg to edge node error: %v, operation is [%s], resource is [%s]",
			err, msg.GetOption(), msg.GetResource())
		return fmt.Errorf("send to client [%s] failed", msg.GetNodeId())
	}

	hwlog.RunLog.Infof("cloud hub send msg to edge node success, operation is [%s], resource is [%s]",
		msg.GetOption(), msg.GetResource())

	return nil
}

func initNodeTable() error {
	if err := database.CreateTableIfNotExist(AuthFailedRecord{}); err != nil {
		hwlog.RunLog.Error("create token failed record database table failed")
		return err
	}
	if err := database.CreateTableIfNotExist(LockRecord{}); err != nil {
		hwlog.RunLog.Error("create token lock database table failed")
		return err
	}
	return nil
}

func doLock() {
	ticker := time.NewTicker(common.LockInterval)
	defer ticker.Stop()

	select {
	case _, ok := <-ticker.C:
		if !ok {
			return
		}
		if err := LockRepositoryInstance().UnlockRecords(); err != nil {
			hwlog.RunLog.Errorf("unlock in db error:%v", err)
			return
		}
		hwlog.RunLog.Info("time expired, automatically unlock token")
		hwlog.OpLog.Infof("[%s@%s] time expired, automatically unlock token",
			constants.MefCenterUserName, constants.LocalHost)
		if !atomic.CompareAndSwapInt64(&lockFlag, lockingFlag, nolockingFlag) {
			hwlog.RunLog.Error("token has unlocked")
		}
	}

}

func checkAndLock() {
	lock, err := LockRepositoryInstance().isLock()
	if err != nil {
		hwlog.RunLog.Error("unlock edge failed")
		return
	}
	if !lock {
		return
	}
	if !atomic.CompareAndSwapInt64(&lockFlag, nolockingFlag, lockingFlag) {
		hwlog.RunLog.Error("token is in locking status, try it later")
		return
	}
	go doLock()
}

func issueCertForEdge(msg *model.Message) (*model.Message, bool, error) {
	var csrBase64Str string
	if err := msg.ParseContent(&csrBase64Str); err != nil {
		hwlog.RunLog.Errorf("parse content failed: %v", err)
		return nil, false, errors.New("parse content failed")
	}
	csrData, err := base64.StdEncoding.DecodeString(csrBase64Str)
	if err != nil {
		hwlog.RunLog.Errorf("decode base64 csr data error: %v", err)
		return nil, false, errors.New("decode base64 csr data error")
	}
	reqCertParams := requests.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath: constants.RootCaPath,
			CertPath:   constants.ServerCertPath,
			KeyPath:    constants.ServerKeyPath,
			SvrFlag:    false,
			WithBackup: true,
		},
	}
	certStr, err := reqCertParams.ReqIssueSvrCert(common.WsCltName, csrData)
	if err != nil {
		hwlog.RunLog.Errorf("issue cert for edge error: %v", err)
		return nil, false, errors.New("issue cert for edge error")
	}
	respMsg, err := msg.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("new response for issuing cert to edge failed, error: %v", err)
		return nil, false, errors.New("new response failed")
	}
	respMsg.SetRouter(
		common.CloudHubName,
		common.EdgeHubName,
		common.OptResp,
		common.ResEdgeCert)
	respMsg.SetNodeId(msg.GetNodeId())
	if err = respMsg.FillContent(certStr); err != nil {
		hwlog.RunLog.Errorf("fill resp into content failed: %v", err)
		return nil, false, errors.New("fill resp into content failed")
	}
	return respMsg, true, nil
}

func (c *CloudServer) getEdgeConnStatus(msg *model.Message) (*model.Message, bool, error) {
	var snList []string
	if err := msg.ParseContent(&snList); err != nil {
		hwlog.RunLog.Errorf("parse content failed: %v", err)
		return nil, false, errors.New("parse content failed")
	}
	var peerInfoList []websocketmgr.WebsocketPeerInfo
	for _, sn := range snList {
		peerInfo, err := c.serverProxy.GetPeer(sn)
		if err != nil {
			hwlog.RunLog.Warnf("failed to get connection status for node %s, %v", sn, err)
			continue
		}
		peerInfoList = append(peerInfoList, *peerInfo)
	}

	msg, err := msg.NewResponse()
	if err != nil {
		return nil, false, fmt.Errorf("failed to create response for edge connection status request, %v", err)
	}
	if err = msg.FillContent(peerInfoList); err != nil {
		return nil, false, fmt.Errorf("failed to fill peer info into content: %v", err)
	}
	if err = modulemgr.SendMessage(msg); err != nil {
		return nil, false, fmt.Errorf("failed to send response for edge connection status request, %v", err)
	}
	return nil, false, nil
}
