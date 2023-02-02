// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package certmanager module cert manager
package certmanager

import (
	"context"
	"errors"

	"edge-manager/pkg/util"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"

	"huawei.com/mindx/common/hwlog"
)

// CertMgr cert manager struct
type CertMgr struct {
	rootCAs RootCAs
	ctx     context.Context
	enable  bool
}

// Name returns the name of cert manager module
func (cm *CertMgr) Name() string {
	return common.CertManagerName
}

// Enable indicates whether this module is enabled
func (cm *CertMgr) Enable() bool {
	return cm.enable
}

// Start generate root cas and issue service cert
func (cm *CertMgr) Start() {
	rootCAs := NewRootCAs()
	rootCAs.GenerateRootCA()

	for {
		select {
		case _, ok := <-cm.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("has listened stop signal")
			return
		default:
		}

		message, err := modulemanager.ReceiveMessage(cm.Name())
		if err != nil {
			hwlog.RunLog.Errorf("receive message from channel failed, error: %v", err)
			continue
		}
		if !util.CheckInnerMsg(message) {
			hwlog.RunLog.Error("message receive from module is invalid")
			continue
		}

		if err = cm.handleIssueService(message); err != nil {
			hwlog.RunLog.Error("issue service cert failed")
			continue
		}
	}
}

// IssueReq issue service cert request struct
type IssueReq struct {
	nodeId  string
	csrByte []byte
}

// ServiceCert service cert struct
type ServiceCert struct {
	NodeId      string `json:"node_id"`
	ServiceCert []byte `json:"service_cert"`
}

func (cm *CertMgr) handleIssueService(message *model.Message) error {
	if message.GetOption() != common.Issue || message.GetDestination() != common.CertManagerName {
		return errors.New("message option or destination is invalid")
	}

	issueReq, ok := message.GetContent().(IssueReq)
	if !ok {
		hwlog.RunLog.Error("convert to issueReq failed")
		return errors.New("convert message content failed")
	}
	serviceCert := cm.rootCAs.issueService(issueReq.csrByte)

	respMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("new message failed, error: %v", err)
		return err
	}
	content := &ServiceCert{
		NodeId:      issueReq.nodeId,
		ServiceCert: serviceCert,
	}
	respMsg.SetRouter(cm.Name(), common.CloudHubName, message.GetOption(), common.ServiceCert)
	respMsg.FillContent(content)
	respMsg.SetIsSync(false)
	if err = modulemanager.SendMessage(respMsg); err != nil {
		hwlog.RunLog.Errorf("send message failed, error: %v", err)
		return err
	}

	return nil
}

// NewCertManager new CertMgr
func NewCertManager(enable bool) *CertMgr {
	socket := &CertMgr{
		enable: enable,
		ctx:    context.Background(),
	}
	return socket
}
