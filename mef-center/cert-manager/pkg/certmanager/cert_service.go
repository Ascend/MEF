// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package certmanager init cert restful service
package certmanager

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509"
	"huawei.com/mindx/common/x509/certutils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"

	"cert-manager/pkg/certmanager/certchecker"
)

const (
	serialNumberLen = 20
	sha256sumLen    = 32
)

var caLock sync.Mutex

func queryRootCa(msg *model.Message) common.RespMsg {
	var certName string
	if err := msg.ParseContent(&certName); err != nil {
		hwlog.RunLog.Errorf("query cert info failed: parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}
	if err := certchecker.CheckCertName(certName); err != nil {
		hwlog.RunLog.Error("the cert name not support")
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: "Query root ca failed", Data: nil}
	}
	if !isCertImported(certName) {
		hwlog.RunLog.Infof("query [%s] cert finished, which is not imported yet", certName)
		return common.RespMsg{Status: common.Success, Msg: fmt.Sprintf("%s cert is no imported yet", certName), Data: ""}
	}
	ca, err := getCertByCertName(certName)
	if err != nil {
		hwlog.RunLog.Errorf("query cert [%s] root ca failed: %v", certName, err)
		return common.RespMsg{Status: common.ErrorGetRootCa, Msg: "Query root ca failed", Data: nil}
	}
	if ca == nil {
		hwlog.RunLog.Errorf("cert [%s] root ca not exist", certName)
		return common.RespMsg{Status: common.ErrorGetRootCa, Msg: "Query root ca file not exist", Data: nil}
	}
	// if cert update is in process, return both old and new hub_client ca certs
	if certName == common.WsCltName {
		tempCaFilePath := getTempRootCaPath(certName)
		if fileutils.IsExist(tempCaFilePath) {
			tempCaBytes, err := certutils.GetCertContent(tempCaFilePath)
			if err != nil {
				hwlog.RunLog.Errorf("load new temp root cert failed: %v, only old root cert will be used", err)
				return common.RespMsg{Status: common.ErrorGetRootCa, Msg: "failed to load new temp root ca", Data: nil}
			}
			ca = append(ca, tempCaBytes...)
		}
	}
	hwlog.RunLog.Infof("query [%s] root ca success", certName)
	return common.RespMsg{Status: common.Success, Msg: "query ca success", Data: string(ca)}
}

func issueServiceCa(msg *model.Message) common.RespMsg {
	var csrJsonData csrJson
	if err := msg.ParseContent(&csrJsonData); err != nil {
		hwlog.RunLog.Errorf("parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}
	if checkResult := certchecker.NewIssueCertChecker().Check(csrJsonData); !checkResult.Result {
		hwlog.RunLog.Errorf("cert issue para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: "cert issue para check failed", Data: nil}
	}
	cert, err := issueServiceCert(csrJsonData.CertName, csrJsonData.Csr)
	if err != nil {
		hwlog.RunLog.Errorf("issue service certificate failed: %v", err)
		return common.RespMsg{Status: common.ErrorIssueSrvCert, Msg: "issue service certificate failed", Data: nil}
	}
	hwlog.RunLog.Infof("issue [%s] service certificate success", csrJsonData.CertName)
	return common.RespMsg{Status: common.Success, Msg: "issue success", Data: string(cert)}
}

func certsUpdateResult(msg *model.Message) common.RespMsg {
	var result certUpdateResult
	if err := msg.ParseContent(&result); err != nil {
		hwlog.RunLog.Errorf("parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed"}
	}
	switch result.CertType {
	case CertTypeEdgeCa:
		if edgeCaResultChan == nil {
			edgeCaResultChan = make(chan certUpdateResult)
		}
		edgeCaResultChan <- result
	case CertTypeEdgeSvc:
		if edgeSvcResultChan == nil {
			edgeSvcResultChan = make(chan certUpdateResult)
		}
		edgeSvcResultChan <- result
	default:
		hwlog.RunLog.Errorf("cert type error: %v", result.CertType)
		return common.RespMsg{Status: common.ErrorCertTypeError, Msg: common.ErrorMap[common.ErrorCertTypeError]}
	}
	return common.RespMsg{Status: common.Success, Msg: ""}
}

func importRootCa(msg *model.Message) common.RespMsg {
	caLock.Lock()
	defer caLock.Unlock()
	hwlog.RunLog.Info("import cert item start")
	var req importCertReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}
	if checkResult := certchecker.NewImportCertChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("cert import para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid,
			Msg: fmt.Sprintf("cert import para check failed: %s", checkResult.Reason)}
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
	certMonitor.Recheck()

	hwlog.RunLog.Infof("import %s certificate success", req.CertName)
	return common.RespMsg{Status: common.Success, Msg: "import certificate success", Data: nil}
}

func deleteRootCa(msg *model.Message) common.RespMsg {
	caLock.Lock()
	defer caLock.Unlock()
	hwlog.RunLog.Info("import the cert item start")
	var req deleteCaReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}
	if checkResult := certchecker.NewDeleteCertChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("cert delete para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid,
			Msg: fmt.Sprintf("cert delete para check failed: %s", checkResult.Reason)}
	}
	// delete root certificate content
	if err := removeCaFile(req.Type); err != nil {
		hwlog.RunLog.Errorf("delete ca file failed, error:%v", err)
		return common.RespMsg{Status: common.ErrorDeleteRootCa, Msg: "delete ca file failed", Data: nil}
	}
	certMonitor.Recheck()

	hwlog.RunLog.Infof("delete %s certificate success", req.Type)
	return common.RespMsg{Status: common.Success, Msg: "delete ca file success", Data: nil}
}

func getCertInfo(msg *model.Message) common.RespMsg {
	var certName string
	if err := msg.ParseContent(&certName); err != nil {
		hwlog.RunLog.Errorf("get cert info failed: parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed"}
	}
	if !certchecker.CheckIfCanGetInfo(certName) {
		hwlog.RunLog.Error("the cert name not support")
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: "the cert name not support"}
	}
	ca, err := getCertByCertName(certName)
	if err != nil {
		hwlog.RunLog.Errorf("load cert [%s] failed, %v", certName, err)
		return common.RespMsg{Status: common.ErrorGetRootCaInfo, Msg: "load cert failed"}
	}
	info, err := parseNorthernRootCa(ca)
	if err != nil {
		hwlog.RunLog.Errorf("cert [%s] root ca parse cert failed: %v", certName, err)
		return common.RespMsg{Status: common.ErrorGetRootCaInfo, Msg: "parse cert failed"}
	}
	return common.RespMsg{Status: common.Success, Data: info}
}

func parseNorthernRootCa(caBytes []byte) (interface{}, error) {
	caChainMgr, err := x509.NewCaChainMgr(caBytes)
	if err != nil {
		hwlog.RunLog.Errorf("create ca chain failed, %v", err)
		return nil, err
	}

	var infos []map[string]interface{}
	for _, cert := range caChainMgr.GetCerts() {
		sha256sum := sha256.Sum256(cert.Raw)
		cInfo := map[string]interface{}{
			"Issuer":       cert.Issuer.String(),
			"Subject":      cert.Subject.String(),
			"SerialNumber": utils.BinaryFormat(cert.SerialNumber.Bytes(), serialNumberLen),
			"Validity": map[string]interface{}{
				"NotBefore": cert.NotBefore.In(time.Local).Format(common.TimeFormat),
				"NotAfter":  cert.NotAfter.In(time.Local).Format(common.TimeFormat),
			},
			"FingerPrintAlgorithm": "sha256",
			"FingerPrint":          utils.BinaryFormat(sha256sum[:], sha256sumLen),
		}
		infos = append(infos, cInfo)
	}

	return infos, nil
}

func importCrl(msg *model.Message) common.RespMsg {
	caLock.Lock()
	defer caLock.Unlock()
	hwlog.RunLog.Info("start to import the crl")
	var req importCrlReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed"}
	}
	if checkResult := certchecker.NewImportCrlChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("import crl para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid,
			Msg: fmt.Sprintf("import crl para check failed, %s", checkResult.Reason)}
	}
	// base64 decode crl content
	bytes, err := base64.StdEncoding.DecodeString(req.Crl)
	if err != nil {
		hwlog.RunLog.Errorf("base64 decode ca content failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamInvalid,
			Msg: fmt.Sprintf("base64 decode ca content failed, error: %v", err)}
	}

	if err := saveCrlContentWithBackup(req.CrlName, bytes); err != nil {
		return common.RespMsg{Status: common.ErrorSaveCrl,
			Msg: fmt.Sprintf("save ca content to file failed, error: %v", err)}
	}
	certMonitor.Recheck()

	return common.RespMsg{Status: common.Success, Msg: "import crl file success"}
}

// saveCaContent save ca content to File
func saveCrlContentWithBackup(crlName string, crlContent []byte) error {
	crlFilePath := getCrlPath(crlName)
	if err := fileutils.MakeSureDir(crlFilePath); err != nil {
		hwlog.RunLog.Errorf("create %s crl folder failed, error: %v", crlName, err)
		return fmt.Errorf("create %s crl folder failed, error: %v", crlName, err)
	}
	if err := fileutils.WriteData(crlFilePath, crlContent); err != nil {
		hwlog.RunLog.Errorf("save %s crl file failed, error:%s", crlName, err)
		return fmt.Errorf("save %s crl file failed", crlName)
	}
	if err := backuputils.BackUpFiles(crlFilePath); err != nil {
		hwlog.RunLog.Errorf("create backup for %s crl file failed, error:%s", crlName, err)
		return fmt.Errorf("create backup for %s crl file failed", crlName)
	}
	hwlog.RunLog.Infof("save %s crl file success", crlName)
	return nil
}

func queryCrl(msg *model.Message) common.RespMsg {
	var crlName string
	if err := msg.ParseContent(&crlName); err != nil {
		hwlog.RunLog.Errorf("query crl info failed: parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}
	if checkResult := certchecker.NewImportCrlNameChecker().Check(crlName); !checkResult.Result {
		hwlog.RunLog.Error("the crl name not support")
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: "query crl failed parma is invalid", Data: nil}
	}
	if !isExternalCrlImported(crlName) {
		hwlog.RunLog.Infof("query [%s] crl finished, which is not imported yet", crlName)
		return common.RespMsg{Status: common.Success, Msg: fmt.Sprintf("%s crl is no imported yet", crlName), Data: ""}
	}
	crlData, err := certutils.GetCrlContentWithBackup(getCrlPath(crlName))
	if err != nil {
		hwlog.RunLog.Errorf("[%s] crl file is damaged", crlName)
		return common.RespMsg{Status: common.ErrorGetRootCa, Msg: "query crl failed, crl file is damaged", Data: nil}
	}
	hwlog.RunLog.Infof("query [%s] crl success", crlName)
	return common.RespMsg{Status: common.Success, Msg: "query crl success", Data: string(crlData)}
}

func getImportedCertsInfo(*model.Message) common.RespMsg {
	hwlog.RunLog.Info("start to get imported certs info")
	resp := requests.ImportedCertsInfo{}

	const contentEmptyStr = "content is empty"
	northCertPath := filepath.Join(util.RootCaMgrDir, common.NorthernCertName, util.RootCaFileName)
	northCertInfo, err := x509.CheckCertsChainReturnContent(northCertPath)
	if err != nil && strings.Contains(err.Error(), contentEmptyStr) {
		hwlog.RunLog.Warn("get north cert info failed, cert is nil")
	} else if err != nil {
		hwlog.RunLog.Errorf("get north cert info failed, error: %v", err)
	}
	if err == nil {
		hwlog.RunLog.Info("get north cert info success")
		resp.NorthCert = northCertInfo
	}

	softwareCertPath := filepath.Join(util.RootCaMgrDir, common.SoftwareCertName, util.RootCaFileName)
	softwareCertInfo, err := x509.CheckCertsChainReturnContent(softwareCertPath)
	if err != nil && strings.Contains(err.Error(), contentEmptyStr) {
		hwlog.RunLog.Warn("get software repository cert info failed, cert is nil")
	} else if err != nil {
		hwlog.RunLog.Errorf("get software repository cert info failed, error: %v", err)
	}
	if err == nil {
		hwlog.RunLog.Info("get software repository cert info success")
		resp.SoftwareCert = softwareCertInfo
	}

	imageCertPath := filepath.Join(util.RootCaMgrDir, common.ImageCertName, util.RootCaFileName)
	imageCertInfo, err := x509.CheckCertsChainReturnContent(imageCertPath)
	if err != nil && strings.Contains(err.Error(), contentEmptyStr) {
		hwlog.RunLog.Warn("get image repository cert info failed, cert is nil")
	} else if err != nil {
		hwlog.RunLog.Errorf("get image repository cert info failed, error: %v", err)
	}
	if err == nil {
		hwlog.RunLog.Info("get image repository cert info success")
		resp.ImageCert = imageCertInfo
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		hwlog.RunLog.Errorf("marshal imported certs info failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorGetImportedCertsInfo, Msg: "get imported certs info failed", Data: nil}
	}

	hwlog.RunLog.Info("get imported certs info success")
	return common.RespMsg{Status: common.Success, Msg: "get imported certs info success", Data: string(respBytes)}
}
