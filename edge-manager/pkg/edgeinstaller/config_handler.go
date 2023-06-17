// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller config handler
package edgeinstaller

import (
	"encoding/json"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-manager/pkg/kubeclient"
	"huawei.com/mindxedge/base/common"
)

type configHandler struct{}

// TokenResp token response from edge-installer
type TokenResp struct {
	Token []byte `json:"token"`
}

// Handle configHandler handle entry
func (ch *configHandler) Handle(message *model.Message) error {
	hwlog.RunLog.Info("edge-installer received message from edge-connector success to get token")

	token, err := kubeclient.GetKubeClient().GetToken()
	defer common.ClearSliceByteMemory(token)
	if err != nil {
		hwlog.RunLog.Errorf("get token from k8s failed, error: %v", err)
		return fmt.Errorf("get token from k8s failed, error: %v", err)
	}

	tokenResp := TokenResp{
		Token: token,
	}
	defer common.ClearSliceByteMemory(tokenResp.Token)
	data, err := json.Marshal(tokenResp)
	if err != nil {
		hwlog.RunLog.Errorf("marshal token response failed, error: %v", err)
		return fmt.Errorf("marshal token response failed, error: %v", err)
	}

	if err = sendMessage(message, string(data)); err != nil {
		hwlog.RunLog.Errorf("edge-installer send message to edge-connector for config failed, error: %v", err)
		return fmt.Errorf("edge-installer send message to edge-connector for config failed, error: %v", err)
	}

	hwlog.RunLog.Info("edge-installer send message to edge-connector success with token")
	return nil
}
