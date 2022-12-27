// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller the edge-installer module related
package edgeinstaller

import (
	"context"
	"encoding/json"
	"errors"

	"edge-manager/pkg/database"
	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/util"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"

	"huawei.com/mindx/common/hwlog"
)

// Installer edge-installer struct
type Installer struct {
	ctx    context.Context
	enable bool
}

// Name returns the name of edge installer module
func (i *Installer) Name() string {
	return common.EdgeInstallerName
}

// Enable indicates whether this module is enabled
func (i *Installer) Enable() bool {
	if i.enable {
		if err := initSoftwareMgrInfoTable(); err != nil {
			hwlog.RunLog.Errorf("module (%s) init database table failed, cannot enable", common.EdgeInstaller)
			return !i.enable
		}
	}

	return i.enable
}

// Start sends and receives message
func (i *Installer) Start() {
	for {
		select {
		case _, ok := <-i.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("has listened stop signal")
			return
		default:
		}

		message, err := modulemanager.ReceiveMessage(i.Name())
		if err != nil {
			hwlog.RunLog.Errorf("receive message from channel failed, error: %v", err)
			continue
		}

		if !util.CheckInnerMsg(message) {
			hwlog.RunLog.Error("message receive from module is invalid")
			continue
		}

		if message.GetDestination() != i.Name() {
			hwlog.RunLog.Errorf("message is not sent to edge-connector, destination is [%s]", message.GetDestination())
			continue
		}

		switch message.GetOption() {
		case common.Upgrade:
			go i.dealUpgrade(message)
		case common.Download:
			go i.dealDownload(message)
		case common.Get:
			go i.dealGetToken(message)
		default:
			hwlog.RunLog.Error("invalid operation")
			continue
		}
	}
}

// TokenReq token request from edge-installer
type TokenReq struct {
	NodeId string `json:"nodeId"`
	Token  []byte `json:"token,omitempty"`
}

func (i *Installer) dealGetToken(message *model.Message) {
	hwlog.RunLog.Info("edge-installer received message from edge-connector success to get token")
	if !(message.GetSource() == common.EdgeConnectorName) || !(message.GetResource() == common.Token) {
		hwlog.RunLog.Error("invalid source or resource")
		return
	}

	data, ok := message.GetContent().(string)
	if !ok {
		hwlog.RunLog.Error("convert to tokenReq failed")
		return
	}

	token, err := kubeclient.GetKubeClient().GetToken()
	if err != nil {
		hwlog.RunLog.Error("get token from k8s failed")
		return
	}

	tokenReq := TokenReq{}
	if err = json.Unmarshal([]byte(data), &tokenReq); err != nil {
		hwlog.RunLog.Errorf("parse data getting token failed, error: %v", err)
		return
	}

	tokenReq.Token = token
	if err = i.sendTokenToEdgeConnector(&tokenReq, message.GetOption()); err != nil {
		hwlog.RunLog.Errorf("send token to edge connector failed, error: %v", err)
		return
	}

	hwlog.RunLog.Info("edge-installer send to edge-connector success with token")
	return
}

func (i *Installer) sendTokenToEdgeConnector(tokenReq *TokenReq, option string) error {
	content := *tokenReq

	sendMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("new message failed, error: %v", err)
		return err
	}
	sendMsg.SetRouter(i.Name(), common.EdgeConnectorName, option, common.Token)
	sendMsg.FillContent(content)
	sendMsg.SetIsSync(false)
	if err = modulemanager.SendMessage(sendMsg); err != nil {
		hwlog.RunLog.Errorf("send message failed, error: %v", err)
		return err
	}

	return nil
}

func (i *Installer) dealUpgrade(message *model.Message) {
	hwlog.RunLog.Info(" ----------deal upgrade request from edge-installer begin-------")
	if !(message.GetSource() == common.RestfulServiceName) || !(message.GetResource() == common.Software) {
		hwlog.RunLog.Error("invalid source or resource")
		return
	}

	if err := i.respRestful(message); err != nil {
		hwlog.RunLog.Error("send response for upgrading to restful module failed")
		return
	}
	hwlog.RunLog.Info("edge-installer send SUCCESS to restful for upgrading success")

	var dealSfwReq util.UpgradeSfwReq
	if err := common.ParamConvert(message.GetContent(), &dealSfwReq); err != nil {
		hwlog.RunLog.Error("convert to dealSfwReq failed")
		return
	}

	dealSfwContent, err := upgradeWithSfwManager(dealSfwReq)
	if err != nil {
		hwlog.RunLog.Error("deal with software manager failed")
		return
	}

	if err = i.sendToEdgeConnector(dealSfwContent, message.GetOption()); err != nil {
		hwlog.RunLog.Errorf("send to edge-connector failed, error: %v", err)
		return
	}

	hwlog.RunLog.Info("edge-installer send to edge-connector success with download url for upgrading")
	return
}

func (i *Installer) respRestful(message *model.Message) error {
	respToRestful, respContent := i.constructContent(message)
	respToRestful.FillContent(respContent)
	if err := modulemanager.SendMessage(respToRestful); err != nil {
		hwlog.RunLog.Errorf("%s send response to restful failed", common.EdgeInstallerName)
		return err
	}

	return nil
}

func (i *Installer) constructContent(message *model.Message) (*model.Message, common.RespMsg) {
	var req util.UpgradeSfwReq
	if err := common.ParamConvert(message.GetContent(), &req); err != nil {
		return nil, common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	respToRestful, err := message.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("%s new response failed", i.Name())
		return nil, common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	return respToRestful, common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func (i *Installer) dealDownload(message *model.Message) {
	hwlog.RunLog.Info("edge-installer received message from edge-connector success")
	if !(message.GetSource() == common.EdgeConnectorName) || !(message.GetResource() == common.Software) {
		hwlog.RunLog.Error("invalid source or resource")
		return
	}

	downloadSfwReq, ok := message.GetContent().(util.DownloadSfwReq)
	if !ok {
		hwlog.RunLog.Error("convert to dealSfwReq failed")
		return
	}

	dealSfwContent, err := downloadWithSfwMgr(downloadSfwReq)
	if err != nil {
		hwlog.RunLog.Error("deal with software manager failed")
		return
	}

	if err = i.sendToEdgeConnector(dealSfwContent, message.GetOption()); err != nil {
		hwlog.RunLog.Errorf("send to edge-connector failed, error: %v", err)
		return
	}

	hwlog.RunLog.Infof("edge-installer send to edge-connector success with download url for downloading [%s]",
		dealSfwContent.SoftwareName)
	return
}

func (i *Installer) sendToEdgeConnector(dealSfwContent *util.DealSfwContent, option string) error {
	content := *dealSfwContent

	sendMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("new message failed, error: %v", err)
		return err
	}
	sendMsg.SetRouter(i.Name(), common.EdgeConnectorName, option, common.Software)
	sendMsg.FillContent(content)
	sendMsg.SetIsSync(false)
	if err = modulemanager.SendMessage(sendMsg); err != nil {
		hwlog.RunLog.Errorf("send message failed, error: %v", err)
		return err
	}

	return nil
}

// NewInstaller new Installer
func NewInstaller(enable bool) *Installer {
	socket := &Installer{
		ctx:    context.Background(),
		enable: enable,
	}
	return socket
}

func initSoftwareMgrInfoTable() error {
	if err := database.CreateTableIfNotExists(SoftwareMgrInfo{}); err != nil {
		return errors.New("table software manager info create failed")
	}

	if err := CreateTableSfwInfo(); err != nil {
		hwlog.RunLog.Error("create item in table software manager info failed")
		return errors.New("create item in table software manager info failed")
	}

	return nil
}
