// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller upgrade handler
package edgeinstaller

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
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
	hwlog.RunLog.Info("edge-installer received message success for upgrading software from restful module")

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

	_, err = getNodeNum(upgradeSfwReq.NodeNums)
	if err != nil {
		hwlog.RunLog.Errorf("get node unique name when upgrading failed, error: %v", err)
		return fmt.Errorf("get node unique name when upgrading failed, error: %v", err)
	}
	return nil
}
