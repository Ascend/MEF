// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager to manage node msg
package edgemsgmanager

import (
	"encoding/json"
	"time"

	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/util"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/certutils"
	"huawei.com/mindxedge/base/common/httpsmgr"
	"huawei.com/mindxedge/base/modulemanager/model"
)

const (
	retryTime = 30
	waitTime  = 5 * time.Second
)

// GetCertInfo [method] get root cert
func GetCertInfo(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("----------downloading cert content begin----------")
	message, ok := input.(*model.Message)
	if !ok {
		hwlog.RunLog.Errorf("get message failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get message failed", Data: nil}
	}

	certName, ok := message.GetContent().(string)
	if !ok {
		hwlog.RunLog.Error("message content type invalid")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "message content type invalid", Data: nil}
	}

	reqCertParams := httpsmgr.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath:    util.RootCaPath,
			CertPath:      util.ServerCertPath,
			KeyPath:       util.ServerKeyPath,
			SvrFlag:       false,
			IgnoreCltCert: false,
		},
	}
	var rootCaRes string
	var err error
	for i := 0; i < retryTime; i++ {
		rootCaRes, err = reqCertParams.GetRootCa(certName)
		if err == nil {
			break
		}
		time.Sleep(waitTime)
	}
	address, err := util.GetImageAddress()
	if err != nil {
		hwlog.RunLog.Errorf("get image registry address failed, %v", err)
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "message content type invalid", Data: nil}
	}
	res := certutils.QueryCertRes{
		CertName: certName,
		Cert:     rootCaRes,
		Address:  address,
	}

	data, err := json.Marshal(res)
	if err != nil {
		hwlog.RunLog.Errorf("marshal cert response failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "message content type invalid", Data: nil}
	}

	if err = sendMessageToEdge(message, string(data)); err != nil {
		hwlog.RunLog.Errorf("edge msg manager send message to edge hub for config failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorSendMsgToNode, Msg: "send msg to edge failed", Data: nil}
	}

	hwlog.RunLog.Info("edge msg manager send message to edge hub success with cert info")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}
