// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller the edge-installer module related
package edgeinstaller

import (
	"context"
	"errors"

	"edge-manager/pkg/database"
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
		default:
			hwlog.RunLog.Error("invalid operation")
			continue
		}
	}
}

func (i *Installer) dealUpgrade(message *model.Message) {
	if !(message.GetSource() == common.RestfulServiceName) || !(message.GetResource() == common.Software) {
		hwlog.RunLog.Error("invalid source or resource")
		return
	}

	if err := i.respRestful(message); err != nil {
		hwlog.RunLog.Error("send response to restful module failed")
		return
	}

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

	hwlog.RunLog.Info("edge-installer send to edge-connector success with download url")
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
