// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package certmanager init cert restful service
package certmanager

import (
	"encoding/base64"

	"huawei.com/mindx/common/hwlog"

	"cert-manager/pkg/certmanager/certchecker"
	"huawei.com/mindxedge/base/common"
)

func queryRootCa(input interface{}) common.RespMsg {
	certName, ok := input.(string)
	if !ok {
		hwlog.RunLog.Error("query cert info failed: para type not valid")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "query cert info request convert error", Data: nil}
	}
	if !certchecker.CheckCertName(certName) {
		hwlog.RunLog.Error("the cert name not support")
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: "Query root ca failed", Data: nil}
	}
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
	if checkResult := certchecker.NewIssueCertChecker().Check(csrJsonData); !checkResult.Result {
		hwlog.RunLog.Errorf("cert issue para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: "cert issue para check failed", Data: nil}
	}
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
	if checkResult := certchecker.NewImportCertChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("cert import para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: "cert import para check failed", Data: nil}
	}
	// base64 decode root certificate content
	caBase64, err := base64.StdEncoding.DecodeString(req.Cert)
	if err != nil {
		hwlog.RunLog.Errorf("base64 decode ca content failed, error:%v", err)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: "base64 decode ca content failed", Data: nil}
	}
	// save the certificate to the local file
	if err := saveCaContent(req.CertName, caBase64); err != nil {
		hwlog.RunLog.Errorf("save ca content to file failed, error:%v", err)
		return common.RespMsg{Status: common.ErrorSaveCa, Msg: "save ca content to file failed", Data: nil}
	}
	go func() {
		if err := updateClientCert(req.CertName, caBase64); err != nil {
			hwlog.RunLog.Errorf("distribute cert file to client failed, error:%v", err)
		}
	}()
	hwlog.OpLog.Infof("import %s certificate success", req.CertName)
	return common.RespMsg{Status: common.Success, Msg: "import certificate success", Data: nil}
}

func deleteRootCa(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("import the cert item start")
	var req deleteCaReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}
	if checkResult := certchecker.NewDeleteCertChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("cert delete para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: "cert delete para check failed", Data: nil}
	}
	// delete root certificate content
	if err := removeCaFile(req.Type); err != nil {
		hwlog.RunLog.Errorf("delete ca file failed, error:%v", err)
		return common.RespMsg{Status: common.ErrorDeleteRootCa, Msg: "delete ca file failed", Data: nil}
	}
	hwlog.RunLog.Infof("delete %s certificate success", req.Type)
	return common.RespMsg{Status: common.Success, Msg: "delete ca file success", Data: nil}
}
