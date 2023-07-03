// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager get cert info
package edgemsgmanager

import (
	"encoding/json"
	"errors"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/x509/certutils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"

	"edge-manager/pkg/util"
)

func queryCertInfo(certName string) (certutils.ClientCertResp, error) {
	res := certutils.ClientCertResp{}
	reqCertParams := requests.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath: util.RootCaPath,
			CertPath:   util.ServerCertPath,
			KeyPath:    util.ServerKeyPath,
			SvrFlag:    false,
		},
	}
	rootCaRes, err := reqCertParams.GetRootCa(certName)
	if err != nil {
		hwlog.RunLog.Errorf("query %s cert content from cert-manager failed, error: %v", certName, err)
		return res, errors.New("query cert content from cert-manager failed")
	}

	res.CertName = certName
	res.CertContent = rootCaRes
	res.CertOpt = common.Update

	if certName == common.ImageCertName {
		address, err := util.GetImageAddress()
		if err != nil {
			hwlog.RunLog.Errorf("get image registry address failed, error:%v", err)
			return res, errors.New("get image registry address failed")
		}
		if address == "" {
			hwlog.RunLog.Warn("image registry address should be configured")
			return res, errors.New("image registry address should be configured")
		}
		res.ImageAddress = address
	}
	return res, nil
}

// GetCertInfo [method] get root cert
func GetCertInfo(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("----------downloading cert content begin----------")
	message, ok := input.(*model.Message)
	if !ok {
		hwlog.RunLog.Error("get message failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get message failed", Data: nil}
	}

	certName, ok := message.GetContent().(string)
	if !ok {
		hwlog.RunLog.Error("message content type invalid")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "message content type invalid", Data: nil}
	}
	if !newCertNameChecker().Check(certName) {
		hwlog.RunLog.Error("the cert name not support")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "query cert name not support", Data: nil}
	}
	certRes, err := queryCertInfo(certName)
	if err != nil {
		hwlog.RunLog.Errorf("query cert from cert manager failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorQueryCrt, Msg: "query cert from cert manager failed", Data: nil}
	}

	data, err := json.Marshal(certRes)
	if err != nil {
		hwlog.RunLog.Errorf("marshal cert response failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "message content type invalid", Data: nil}
	}

	if err = sendMessageToEdge(message, string(data)); err != nil {
		hwlog.RunLog.Errorf("edge msg manager send message to edge hub with cert info failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorSendMsgToNode, Msg: "send msg to edge failed", Data: nil}
	}

	hwlog.RunLog.Info("edge msg manager send message to edge hub success with cert info")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}
