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

	"edge-manager/pkg/config"
	"edge-manager/pkg/util"
)

func getCertInfo(certName string) (certutils.ClientCertResp, error) {
	res := certutils.ClientCertResp{}

	res.CertName = certName
	certStr, err := config.GetCertCache(certName)
	if err != nil {
		hwlog.RunLog.Errorf("get %s failed, %v", certName, err)
		return res, err
	}
	res.CertContent = certStr
	res.CertOpt = common.Update

	if certName == common.ImageCertName {
		address, err := util.GetImageAddress()
		if err != nil {
			hwlog.RunLog.Errorf("get image registry address failed, error:%v", err)
			return res, errors.New("get image registry address failed")
		}
		res.ImageAddress = address
	}
	return res, nil
}

// GetCertInfo [method] get root cert in response of cert request from mef-edge
func GetCertInfo(msg *model.Message) common.RespMsg {
	var certName string
	if err := msg.ParseContent(&certName); err != nil {
		hwlog.RunLog.Errorf("parse message content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}
	if !newCertNameChecker().Check(certName) {
		hwlog.RunLog.Error("the cert name not support")
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: "query cert name not support", Data: nil}
	}
	hwlog.RunLog.Infof("start send cert[%s] to edge-node", certName)
	// empty configuration can be delivered to edge for notifying that cert has been imported in center.
	certRes, err := getCertInfo(certName)
	if err != nil {
		hwlog.RunLog.Errorf("query cert from cert manager failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorQueryCrt, Msg: "query cert from cert manager failed", Data: nil}
	}

	data, err := json.Marshal(certRes)
	if err != nil {
		hwlog.RunLog.Errorf("marshal cert response failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "message content type invalid", Data: nil}
	}

	if err = sendMessageToEdge(msg, string(data)); err != nil {
		hwlog.RunLog.Errorf("edge msg manager send message to edge hub with cert info failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorSendMsgToNode, Msg: "send msg to edge failed", Data: nil}
	}

	hwlog.RunLog.Info("edge msg manager send message to edge hub success with cert info")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}
