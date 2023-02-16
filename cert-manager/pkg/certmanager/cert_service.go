// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package certmanager init cert restful service
package certmanager

import (
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

func queryRootCa(input interface{}) common.RespMsg {
	certName, ok := input.(string)
	if !ok {
		hwlog.RunLog.Error("query cert info failed: para type not valid")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "query cert info request convert error", Data: nil}
	}
	// todo 增加checker 验证
	ca, err := getCertByCertName(certName)
	if err != nil {
		hwlog.RunLog.Errorf("query cert [%s] root ca failed: %v", certName, err)
		return common.RespMsg{Status: common.ErrorGetRootCa, Msg: "Query root ca failed", Data: nil}
	}
	hwlog.RunLog.Infof("query [%s] root ca success", certName)
	return common.RespMsg{Status: common.Success, Msg: "query ca success", Data: string(ca)}
}

func issueServiceCa(input interface{}) common.RespMsg {
	var csrJsonData csrJson
	if err := common.ParamConvert(input, &csrJsonData); err != nil {
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}
	// todo 增加checker 验证
	cert, err := issueServiceCert(csrJsonData.CertName, csrJsonData.Csr)
	if err != nil {
		hwlog.RunLog.Errorf("issue service certificate failed: %v", err)
		return common.RespMsg{Status: common.ErrorIssueSrvCert, Msg: "issue service certificate failed", Data: string(cert)}
	}
	hwlog.RunLog.Infof("issue [%s] service certificate success", csrJsonData.CertName)
	return common.RespMsg{Status: common.Success, Msg: "issue success", Data: string(cert)}
}

func importRootCa(input interface{}) common.RespMsg {
	hwlog.OpLog.Info("import cert item start")
	var req importCertReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}
	caBase64, err := checkCert(req)
	if err != nil {
		hwlog.OpLog.Error("valid ca content failed, error:%v", err)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: "valid ca content failed", Data: nil}
	}
	// save the certificate to the local file
	if err := ca2File(req.CertName, caBase64); err != nil {
		hwlog.OpLog.Error("save ca content to file failed, error:%v", err)
		return common.RespMsg{Status: common.ErrorSaveCa, Msg: "save ca content to file failed", Data: nil}
	}
	hwlog.OpLog.Infof("import %s certificate success", req.CertName)
	return common.RespMsg{Status: common.Success, Msg: "import certificate success", Data: nil}
}

func queryAlert(input interface{}) common.RespMsg {
	// todo 待实现
	var alertList = [...]string{"alert 1"}
	return common.RespMsg{Status: common.Success, Msg: "query alert success", Data: alertList}
}
