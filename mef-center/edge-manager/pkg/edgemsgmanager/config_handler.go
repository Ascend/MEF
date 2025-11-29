// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgemsgmanager get cloud core token and ca and send to edge
package edgemsgmanager

import (
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

func getToken() ([]byte, error) {
	token, err := kubeclient.GetKubeClient().GetToken()
	if err != nil {
		return nil, err
	}

	if len(token) == 0 || len(token) > maxTokenLen {
		return nil, fmt.Errorf("token len: %d is invalid", len(token))
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

// GetConfigInfo get cloud core token and ca and send to edge
func GetConfigInfo(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("receive msg to get config info")
	var cfgType string
	if err := msg.ParseContent(&cfgType); err != nil {
		hwlog.RunLog.Errorf("parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}

	var dealers = map[string]msgDealer{
		"cloud-core-token": {getToken, utils.ClearSliceByteMemory},
		"cloud-core-ca":    {getCloudCoreCa, nil},
	}

	var (
		dealer msgDealer
		ok     bool
	)
	if dealer, ok = dealers[cfgType]; !ok {
		hwlog.RunLog.Error("config type not support")
		return common.RespMsg{Status: common.ErrorGetConfigData, Msg: "config type not support", Data: nil}
	}

	data, err := dealer.deal()
	if err != nil {
		hwlog.RunLog.Errorf("get config data failed: %v", err)
		return common.RespMsg{Status: common.ErrorGetConfigData, Msg: "get config data failed", Data: nil}
	}

	if dealer.post != nil {
		defer dealer.post(data)
	}

	if err = sendRespToEdge(msg, string(data)); err != nil {
		hwlog.RunLog.Errorf("send msg to edge failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorSendMsgToNode, Msg: "send msg to edge failed", Data: nil}
	}

	hwlog.RunLog.Info("send msg to edge success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}
