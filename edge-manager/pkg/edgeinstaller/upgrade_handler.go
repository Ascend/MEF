// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller upgrade handler
package edgeinstaller

import (
	"encoding/json"
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager/model"
)

type upgradeHandler struct{}

// UpgradeSfwReq upgrade software request from restful
type UpgradeSfwReq struct {
	NodeNums            []int64 `json:"nodeNums"`
	SoftwareName        string  `json:"softwareName"`
	SoftwareVersion     string  `json:"softwareVersion"`
	DownloadUrlFromUser string  `json:"downloadUrlFromUser,omitempty"`
	Username            string  `json:"username,omitempty"`
	Password            string  `json:"password,omitempty"`
}

// Handle configHandler handle entry
func (uh *upgradeHandler) Handle(message *model.Message) error {
	hwlog.RunLog.Info("edge-installer received message from restful module success")

	var upgradeSfwReq *UpgradeSfwReq
	var err error
	if upgradeSfwReq, err = respRestful(message); err != nil {
		hwlog.RunLog.Errorf("send response for upgrading to restful module failed, error: %v", err)
		return fmt.Errorf("send response for upgrading to restful module failed, error: %v", err)
	}

	if upgradeSfwReq == nil {
		hwlog.RunLog.Error("upgradeSfwReq is nil")
		return errors.New("upgradeSfwReq is nil")
	}
	hwlog.RunLog.Info("edge-installer send SUCCESS to restful module for upgrading success")

	nodeIds, err := getNodeNum(upgradeSfwReq.NodeNums)
	if err != nil {
		hwlog.RunLog.Errorf("get node unique name when upgrading failed, error: %v", err)
		return fmt.Errorf("get node unique name when upgrading failed, error: %v", err)
	}

	for _, nodeId := range nodeIds {
		if upgradeSfwReq.DownloadUrlFromUser == "" {
			go uh.dealUpgradeSfwReq(message, upgradeSfwReq, nodeId)
		} else {
			go uh.dealUpgradeSfwReqWithUrl(message, upgradeSfwReq, nodeId)
		}
	}

	return nil
}

func (uh *upgradeHandler) dealUpgradeSfwReq(message *model.Message, upgradeSfwReq *UpgradeSfwReq, nodeId string) {
	hwlog.RunLog.Infof("--------edge-installer [%s] upgrade software [%s] begin--------",
		nodeId, upgradeSfwReq.SoftwareName)
	upgradeSfwReqToSfwMgr := DownloadSfwReqToSfwMgr{
		SoftwareName:    upgradeSfwReq.SoftwareName,
		SoftwareVersion: upgradeSfwReq.SoftwareVersion,
	}

	contentToConnector, err := downloadWithSfwMgr(nodeId, upgradeSfwReqToSfwMgr)
	if err != nil {
		hwlog.RunLog.Errorf("get download url failed from software manager, error:%v", err)
		return
	}

	if err = uh.constructMsgSendToConnector(nodeId, message, contentToConnector); err != nil {
		hwlog.RunLog.Errorf("construct message and send to edge-connector failed, error: %v", err)
		return
	}

	hwlog.RunLog.Infof("--------edge-installer [%s] upgrade software [%s] end--------",
		nodeId, upgradeSfwReq.SoftwareName)
	return
}

func (uh *upgradeHandler) dealUpgradeSfwReqWithUrl(message *model.Message,
	upgradeSfwReqWithUrl *UpgradeSfwReq, nodeId string) {
	defer common.ClearStringMemory(upgradeSfwReqWithUrl.Password)
	hwlog.RunLog.Infof("--------edge-installer [%s] upgrade software [%s] with download url begin--------",
		nodeId, upgradeSfwReqWithUrl.SoftwareName)
	contentToConnector := getContentToConnector(upgradeSfwReqWithUrl)

	if err := uh.constructMsgSendToConnector(nodeId, message, contentToConnector); err != nil {
		hwlog.RunLog.Errorf("construct message and send to edge-connector failed, error: %v", err)
		return
	}

	hwlog.RunLog.Infof("--------edge-installer [%s] upgrade software [%s] with download url end--------",
		nodeId, upgradeSfwReqWithUrl.SoftwareName)
	return
}

func (uh *upgradeHandler) constructMsgSendToConnector(nodeId string, message *model.Message,
	contentToConnector *ContentToConnector) error {
	contentToConnectorAfterMarshal, err := json.Marshal(contentToConnector)
	if err != nil {
		hwlog.RunLog.Errorf("marshal content to edge-connector for upgrading failed, error: %v", err)
		return fmt.Errorf("marshal content to edge-connector for upgrading failed, error: %v", err)
	}

	message.SetNodeId(nodeId)
	if err = sendMessage(message, string(contentToConnectorAfterMarshal)); err != nil {
		hwlog.RunLog.Errorf("edge-installer send message to edge-connector for upgrading failed, error: %v", err)
		return fmt.Errorf("edge-installer send message to edge-connector for upgrading failed, error: %v", err)
	}

	return nil
}
