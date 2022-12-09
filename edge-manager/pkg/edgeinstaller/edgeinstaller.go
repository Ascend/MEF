// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller the edge-installer module related
package edgeinstaller

import (
	"context"
	"encoding/json"
	"time"

	"edge-manager/pkg/util"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"

	"huawei.com/mindx/common/hwlog"
)

// WaitSfwSyncTime waiting for a response from the software repository
const WaitSfwSyncTime = 10 * time.Second

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
		msg, err := modulemanager.ReceiveMessage(i.Name())
		if err != nil {
			return
		}
		if !util.CheckInnerMsg(msg) {
			hwlog.RunLog.Error("message receive from module is invalid")
			continue
		}

		respRestful, err := msg.NewResponse()
		if err != nil {
			hwlog.RunLog.Errorf("%s new response failed", common.EdgeInstallerName)
			continue
		}
		respRestful.FillContent(common.RespMsg{Status: common.Success, Msg: "", Data: nil})
		if err = modulemanager.SendMessage(respRestful); err != nil {
			hwlog.RunLog.Errorf("%s send response failed", common.EdgeInstallerName)
			continue
		}

		resp := i.sendSyncToModule(msg)
		mergeContentAndSend(msg, resp)
	}
}

func (i *Installer) sendSyncToModule(msg *model.Message) *model.Message {
	destination := ""
	switch msg.GetOption() {
	case common.Upgrade:
		destination = common.SoftwareManagerName
	default:
		hwlog.RunLog.Error("message destination invalid")
		return nil
	}

	sendMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("new message failed, error: %v", err)
		return nil
	}
	sendMsg.SetRouter(common.EdgeInstallerName, destination, common.Query, common.Software)
	sendMsg.SetIsSync(true)
	resp, err := modulemanager.SendSyncMessage(sendMsg, WaitSfwSyncTime)
	if err != nil {
		hwlog.RunLog.Errorf("wait sync message failed, error: %v", err)
		return nil
	}
	isResponse := i.isSyncResponse(resp.GetParentId())
	if !isResponse {
		hwlog.RunLog.Error("error sync response")
		return nil
	}
	return resp
}

func mergeContentAndSend(msg, resp *model.Message) {
	data, err := json.Marshal(msg.GetContent())
	if err != nil {
		hwlog.RunLog.Errorf("marshal message content failed, error: %v", err)
		return
	}
	respData, err := json.Marshal(resp.GetContent())
	if err != nil {
		hwlog.RunLog.Errorf("marshal resp message content failed, error: %v", err)
		return
	}
	content := make(map[string]interface{})
	if err = json.Unmarshal(data, &content); err != nil {
		hwlog.RunLog.Errorf("parse data failed, error: %v", err)
		return
	}
	if err = json.Unmarshal(respData, &content); err != nil {
		hwlog.RunLog.Errorf("parse resp data failed, error: %v", err)
		return
	}

	respMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("new message failed, error: %v", err)
		return
	}
	respMsg.SetRouter(common.EdgeInstallerName, common.EdgeConnectorName, common.Upgrade, common.Software)
	respMsg.FillContent(content)
	respMsg.SetIsSync(false)
	if err = modulemanager.SendMessage(respMsg); err != nil {
		hwlog.RunLog.Errorf("send message failed, error: %v", err)
		return
	}
}

func (i *Installer) isSyncResponse(msgID string) bool {
	return msgID != ""
}

// NewInstaller new Installer
func NewInstaller(enable bool) *Installer {
	socket := &Installer{
		ctx:    context.Background(),
		enable: enable,
	}
	return socket
}
