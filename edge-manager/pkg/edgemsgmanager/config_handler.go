// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgemsgmanager config handler
package edgemsgmanager

import (
	"encoding/json"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509/certutils"
	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/kubeclient"
)

const maxTokenLen = 1024

// TokenResp token response from edge msg manager
type TokenResp struct {
	Token []byte `json:"token"`
}

type msgDealer struct {
	deal func() ([]byte, error)
	post func([]byte)
}

// GetEdgeConfigInfo get config info
func GetEdgeConfigInfo(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("edge msg manager received message from edge hub success to get token")
	message, ok := input.(*model.Message)
	if !ok {
		hwlog.RunLog.Error("get message failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get message failed", Data: nil}
	}

	token, err := kubeclient.GetKubeClient().GetToken()
	defer common.ClearSliceByteMemory(token)
	if err != nil {
		hwlog.RunLog.Errorf("get token from k8s failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorGetConfigData, Msg: "get token from k8s failed", Data: nil}
	}

	if len(token) == 0 || len(token) > maxTokenLen {
		hwlog.RunLog.Errorf("token len:%d invalid", len(token))
		return common.RespMsg{Status: common.ErrorGetConfigData, Msg: "token len invalid", Data: nil}
	}
	tokenResp := TokenResp{
		Token: token,
	}
	defer common.ClearSliceByteMemory(tokenResp.Token)
	data, err := json.Marshal(tokenResp)
	if err != nil {
		hwlog.RunLog.Errorf("marshal token response failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorGetConfigData, Msg: "get token from k8s failed", Data: nil}
	}

	if err = sendMessageToEdge(message, string(data)); err != nil {
		hwlog.RunLog.Errorf("edge msg manager send message to edge hub for config failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorSendMsgToNode, Msg: "get token from k8s failed", Data: nil}
	}

	hwlog.RunLog.Info("edge msg manager send message to edge hub success with token")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func getToken() ([]byte, error) {
	token, err := kubeclient.GetKubeClient().GetToken()
	if err != nil {
		return nil, err
	}

	if len(token) == 0 || len(token) > maxTokenLen {
		return nil, fmt.Errorf("token len:%d invalid", len(token))
	}

	return token, nil
}

func getCloudCoreCa() ([]byte, error) {
	data, err := kubeclient.GetKubeClient().GetCloudCoreCa()
	if err != nil {
		return nil, err
	}

	return certutils.PemWrapCert(data), nil
}

// GetConfigInfo get config info
func GetConfigInfo(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("receive msg to get config info")
	message, ok := input.(*model.Message)
	if !ok {
		hwlog.RunLog.Error("get message failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get message failed", Data: nil}
	}

	cfgType, ok := message.GetContent().(string)
	if !ok {
		hwlog.RunLog.Error("msg type is not string")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "msg type is not string", Data: nil}
	}

	var dealers = map[string]msgDealer{
		"cloud-core-token": {getToken, utils.ClearSliceByteMemory},
		"cloud-core-ca":    {getCloudCoreCa, nil},
	}

	var dealer msgDealer
	if dealer, ok = dealers[cfgType]; !ok {
		hwlog.RunLog.Errorf("config type not support")
		return common.RespMsg{Status: common.ErrorGetConfigData, Msg: "config type not support", Data: nil}
	}

	data, err := dealer.deal()
	if err != nil {
		hwlog.RunLog.Errorf("get config data failed:%v", err)
		return common.RespMsg{Status: common.ErrorGetConfigData, Msg: "get config data failed", Data: nil}
	}

	if dealer.post != nil {
		defer dealer.post(data)
	}

	if err = sendRespToEdge(message, string(data)); err != nil {
		hwlog.RunLog.Errorf("send msg to edge failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorSendMsgToNode, Msg: "send msg to edge failed", Data: nil}
	}

	hwlog.RunLog.Info("send msg to edge success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}
