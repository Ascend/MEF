// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package cloudhub module edge-connector init
package cloudhub

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/websocketmgr"
	"huawei.com/mindx/common/x509/certutils"

	"edge-manager/pkg/constants"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
)

type messageHandler func(*model.Message) (*model.Message, bool, error)

var messageHandlerMap = make(map[string]messageHandler)

func getMsgHandler(msg *model.Message) (messageHandler, bool) {
	handlerKey := msg.GetOption() + msg.GetResource()
	handler, ok := messageHandlerMap[handlerKey]
	return handler, ok
}
func initMsgHandler() {
	messageHandlerMap[common.OptPost+common.ResEdgeCert] = issueCertForEdge
}

// CloudServer wraps the struct WebSocketServer
type CloudServer struct {
	serverIp     string
	wsPort       int
	authPort     int
	maxClientNum int
	serverProxy  *websocketmgr.WsServerProxy
	writeLock    sync.RWMutex
	ctx          context.Context
	enable       bool
}

var server CloudServer

// NewCloudServer new cloud server
func NewCloudServer(enable bool, wsPort, authPort, maxClientNum int) *CloudServer {
	server = CloudServer{
		wsPort:       wsPort,
		authPort:     authPort,
		maxClientNum: maxClientNum,
		ctx:          context.Background(),
		enable:       enable,
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
	if err := websocketmgr.InitConnLimiter(c.maxClientNum); err != nil {
		hwlog.RunLog.Errorf("init mef edge max client num failed: %v", err)
		return
	}
	go periodCheck()
	var err error
	c.serverProxy, err = InitServer()
	if err != nil {
		hwlog.RunLog.Errorf("init websocket server failed: %v", err)
		return
	}
	initMsgHandler()
	hwlog.RunLog.Info("init websocket server success")
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
		sn := message.GetNodeId()
		ip := message.GetIp()

		hwlog.RunLog.Infof("%v[%v:%v] %v %v start", time.Now().Format(time.RFC3339Nano),
			ip, sn, message.GetOption(), message.GetResource())
		result, sentToEdge, err := handler(message)
		if err != nil {
			hwlog.RunLog.Errorf("%v [%v:%v] %v %v failed", time.Now().Format(time.RFC3339Nano),
				ip, sn, message.GetOption(), message.GetResource())
			hwlog.RunLog.Errorf("process message [option: %v res: %v]error: %v",
				message.GetOption(), message.GetResource(), err)
			return
		}

		hwlog.RunLog.Infof("%v [%v:%v] %v %v success", time.Now().Format(time.RFC3339Nano),
			ip, sn, message.GetOption(), message.GetResource())
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

	resp.FillContent(content)
	if err = modulemgr.SendMessage(resp); err != nil {
		hwlog.RunLog.Errorf("%s send response failed: %v", c.Name(), err)
	}
}

func (c *CloudServer) sendToEdge(msg *model.Message) error {
	sender, err := GetSvrSender()
	if err != nil {
		hwlog.RunLog.Errorf("send to client [%s] failed", msg.GetNodeId())
		return fmt.Errorf("send to client [%s] failed", msg.GetNodeId())
	}

	originalSync := msg.GetIsSync()
	msg.SetIsSync(false)
	defer msg.SetIsSync(originalSync)

	if err = sender.Send(msg.GetNodeId(), msg); err != nil {
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

func periodCheck() {
	unlockIP()
	ticker := time.NewTicker(common.CheckUnlockInterval)
	defer ticker.Stop()
	for {
		select {
		case _, ok := <-ticker.C:
			if !ok {
				return
			}
			unlockIP()
		}
	}
}

func unlockIP() {
	rowsAffected, err := LockRepositoryInstance().UnlockRecords()
	if err != nil {
		hwlog.RunLog.Error("unlock edge failed")
		return
	}
	if rowsAffected == 0 {
		return
	}
	hwlog.RunLog.Info("time expired, automatically unlock token")
	hwlog.OpLog.Infof("[%s@%s] time expired, automatically unlock token", constants.MefCenterUserName, constants.LocalHost)
}

func issueCertForEdge(msg *model.Message) (*model.Message, bool, error) {
	csrRawData := msg.GetContent()
	csrBase64Str, ok := csrRawData.(string)
	if !ok {
		return nil, false, errors.New("csr data format error")
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
		return nil, false, err
	}
	respMsg.SetRouter(
		common.CloudHubName,
		common.EdgeHubName,
		common.OptResp,
		common.ResEdgeCert)
	respMsg.SetNodeId(msg.GetNodeId())
	respMsg.FillContent(certStr)
	return respMsg, true, nil
}
