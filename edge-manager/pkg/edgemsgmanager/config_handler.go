// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgemsgmanager config handler
package edgemsgmanager

import (
	"encoding/json"

	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/kubeclient"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager/model"
)

type configHandler struct{}

// TokenResp token response from edge msg manager
type TokenResp struct {
	Token []byte `json:"token"`
}

// GetConfigInfo get config info
func GetConfigInfo(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("edge msg manager received message from edge hub success to get token")
	message, ok := input.(*model.Message)
	if !ok {
		hwlog.RunLog.Errorf("get message failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get message failed", Data: nil}
	}

	token, err := kubeclient.GetKubeClient().GetToken()
	defer common.ClearSliceByteMemory(token)
	if err != nil {
		hwlog.RunLog.Errorf("get token from k8s failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorGetToken, Msg: "get token from k8s failed", Data: nil}
	}

	tokenResp := TokenResp{
		Token: token,
	}
	defer common.ClearSliceByteMemory(tokenResp.Token)
	data, err := json.Marshal(tokenResp)
	if err != nil {
		hwlog.RunLog.Errorf("marshal token response failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorGetToken, Msg: "get token from k8s failed", Data: nil}
	}

	if err = sendMessageToEdge(message, string(data)); err != nil {
		hwlog.RunLog.Errorf("edge msg manager send message to edge hub for config failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorSendMsgToNode, Msg: "get token from k8s failed", Data: nil}
	}

	hwlog.RunLog.Info("edge msg manager send message to edge hub success with token")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}
