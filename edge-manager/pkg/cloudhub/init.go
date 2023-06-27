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

	"edge-manager/pkg/util"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
)

// CloudServer wraps the struct WebSocketServer
type CloudServer struct {
	wsPort       int
	authPort     int
	maxClientNum int
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
	if err := InitServer(); err != nil {
		hwlog.RunLog.Errorf("init websocket server failed: %v", err)
		return
	}

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
	var err error
	message, err = certProcess(message)
	if err != nil {
		hwlog.RunLog.Errorf("certProcess error: %v", err)
		return
	}
	if c.sendToClient(message) != nil {
		c.response(message, common.FAIL)
	} else {
		c.response(message, common.OK)
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
		hwlog.RunLog.Errorf("%s send response failed", c.Name())
	}
}

func (c *CloudServer) sendToClient(msg *model.Message) error {
	sender, err := GetSvrSender()
	if err != nil {
		hwlog.RunLog.Errorf("send to client [%s] failed", msg.GetNodeId())
		return fmt.Errorf("send to client [%s] failed", msg.GetNodeId())
	}
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
	records, err := LockRepositoryInstance().findUnlockRecords()
	if err != nil {
		hwlog.RunLog.Error(err)
		return
	}
	for _, record := range records {
		if err := LockRepositoryInstance().deleteOneLockRecord(record.IP); err != nil {
			hwlog.RunLog.Warnf("unlock edge(%s) failed", record.IP)
			continue
		}
		hwlog.OpLog.Infof("edge (%s) is unlock", record.IP)
	}
}

func certProcess(msg *model.Message) (*model.Message, error) {
	if isReqCertFromEdge(msg) {
		return issueCertForEdge(msg)
	}
	return msg, nil
}
func isReqCertFromEdge(message *model.Message) bool {
	if message == nil {
		return false
	}
	return message.GetOption() == common.OptPost && message.GetResource() == common.ResEdgeCert
}

func issueCertForEdge(msg *model.Message) (*model.Message, error) {
	csrRawData := msg.GetContent()
	csrBase64Str, ok := csrRawData.(string)
	if !ok {
		return nil, errors.New("csr data format error")
	}
	csrData, err := base64.StdEncoding.DecodeString(csrBase64Str)
	if err != nil {
		hwlog.RunLog.Errorf("decode base64 csr data error: %v", err)
		return nil, errors.New("decode base64 csr data error")
	}
	reqCertParams := requests.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath: util.RootCaPath,
			CertPath:   util.ServerCertPath,
			KeyPath:    util.ServerKeyPath,
			SvrFlag:    false,
		},
	}
	certStr, err := reqCertParams.ReqIssueSvrCert(common.WsCltName, csrData)
	if err != nil {
		hwlog.RunLog.Errorf("issue cert for edge error: %v", err)
		return nil, errors.New("issue cert for edge error")
	}
	respMsg, err := msg.NewResponse()
	if err != nil {
		return nil, err
	}
	respMsg.SetRouter(
		common.CloudHubName,
		common.EdgeHubName,
		common.OptResp,
		common.ResEdgeCert)
	respMsg.SetNodeId(msg.GetNodeId())
	respMsg.FillContent(certStr)
	return respMsg, nil
}
