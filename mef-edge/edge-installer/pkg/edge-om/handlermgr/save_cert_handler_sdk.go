// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

package handlermgr

import (
	"errors"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/x509"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-om/common/certsrequest"
	"edge-installer/pkg/installer/edgectl/imageconfig"
)

// saveCertHandlerSdk config edge core handler
type saveCertHandlerSdk struct {
	res util.ClientCertResp
}

// Handle saveCertHandlerSdk handle entry
func (ch *saveCertHandlerSdk) Handle(*model.Message) error {
	hwlog.RunLog.Info("start to handle update edge cert")
	if err := ch.processCert(); err != nil {
		return err
	}
	hwlog.RunLog.Info("handle update edge cert success")
	return nil
}

func (ch *saveCertHandlerSdk) processCert() error {
	hwlog.RunLog.Info("----------begin deal cert response ----------")
	switch ch.res.CertOpt {
	case constants.OptUpdate:
		if ch.res.CertContent == "" {
			hwlog.RunLog.Infof("%s cert is not imported yet in mef-center", ch.res.CertName)
		} else if err := updateCert(ch.res); err != nil {
			hwlog.RunLog.Errorf("update %s cert failed, error: %v", ch.res.CertName, err)
			return err
		}
		certsrequest.SetReceivedCert(ch.res.CertName)
		// when empty crl data received, skip following operations.
		if len(ch.res.CrlContent) == 0 {
			break
		}
		err := checkCrl(ch.res)
		if err != nil && err != x509.ErrCrlCertNotMatch {
			hwlog.RunLog.Errorf("check crl failed: %v, crl will be discard", err)
			return errors.New("check crl failed, crl will be discard")
		}
		// When both CA and CRL are valid but not match, we treat it as normal situation, only print a warning,
		// then continue to import CA and CRL.
		if err == x509.ErrCrlCertNotMatch {
			hwlog.RunLog.Warnf("crl and cert not match, continue to import")
		}
		if err := saveCrl(ch.res); err != nil {
			hwlog.RunLog.Errorf("save crl for %v failed: %v", ch.res.CertName, err)
			return fmt.Errorf("save crl for %v failed", ch.res.CertName)
		}
		hwlog.RunLog.Infof("process crl for %v success", ch.res.CertName)
	case constants.OptDelete:
		if err := deleteCert(ch.res); err != nil {
			hwlog.RunLog.Errorf("delete %s cert failed, error: %v", ch.res.CertName, err)
			return err
		}
	default:
		hwlog.RunLog.Errorf("cert %s operation not support", ch.res.CertOpt)
		return errors.New("cert operation not support")
	}

	hwlog.RunLog.Infof("process %s cert success", ch.res.CertName)
	return nil
}

func updateCert(res util.ClientCertResp) error {
	var err error
	if err = saveCaContent(res); err != nil {
		hwlog.RunLog.Errorf("save cert content failed, err:%v", err)
		return err
	}
	if res.CertName != constants.ImageCertName || res.ImageAddress == "" {
		return nil
	}

	if err = copyCertToDockerDir(res); err != nil {
		hwlog.RunLog.Errorf("create image cert for docker failed, error: %v", err)
		return err
	}
	flow := imageconfig.NewImageCfgFlow(res.ImageAddress)
	if err = flow.RunTasks(); err == nil {
		return nil
	}

	if deleteErr := util.DeleteImageCertFile(res.ImageAddress); deleteErr != nil {
		hwlog.RunLog.Errorf("delete image config %s failed, error:%v", res.ImageAddress, deleteErr)
	}

	hwlog.RunLog.Errorf("image mapping config %s failed, error: %v", res.ImageAddress, err)
	return fmt.Errorf("image mapping config %s failed, error: %v", res.ImageAddress, err)
}

func deleteCert(res util.ClientCertResp) error {
	// delete root ca
	certDir, err := path.GetCompSpecificDir(constants.ConfigCertPathName)
	if err != nil {
		hwlog.RunLog.Errorf("get cert dir failed, error: %v", err)
		return errors.New("get cert dir failed")
	}
	certFilePath := filepath.Join(certDir, res.CertName, constants.RootCaName)
	backupPath := certFilePath + backuputils.BackupSuffix
	caFilePaths := []string{certFilePath, backupPath}
	deleteSuccess := true
	for _, caFilePath := range caFilePaths {
		if err := fileutils.DeleteFile(caFilePath); err != nil {
			hwlog.RunLog.Errorf("remove %s ca file failed, error: %v", filepath.Base(caFilePath), err)
			deleteSuccess = false
		}
	}
	var deleteErr error
	if !deleteSuccess {
		deleteErr = fmt.Errorf("delete %s ca file failed", res.CertName)
	}
	if res.CertName != constants.ImageCertName || res.ImageAddress == "" {
		return deleteErr
	}
	if err = util.DeleteImageCertFile(res.ImageAddress); err != nil {
		hwlog.RunLog.Errorf("delete image cert file failed, error: %v", err)
		return errors.New("delete image cert failed")
	}
	return deleteErr
}

func saveCaContent(res util.ClientCertResp) error {
	certDir, err := path.GetCompSpecificDir(constants.ConfigCertPathName)
	if err != nil {
		hwlog.RunLog.Errorf("get cert dir failed, error: %v", err)
		return errors.New("get cert dir failed")
	}
	caFilePath := filepath.Join(certDir, res.CertName, constants.RootCaName)
	if err := fileutils.MakeSureDir(caFilePath); err != nil {
		hwlog.RunLog.Errorf("create %s ca folder failed, error: %v", filepath.Base(caFilePath), err)
		return errors.New("create ca folder failed")
	}

	if err := fileutils.WriteData(caFilePath, []byte(res.CertContent)); err != nil {
		hwlog.RunLog.Errorf("update %s cert file failed, error: %v", res.CertName, err)
		return errors.New("update ca file failed")
	}

	if err := backuputils.BackUpFiles(caFilePath); err != nil {
		hwlog.RunLog.Warnf("create backup for %s cert failed, error: %v", res.CertName, err)
	}

	hwlog.RunLog.Infof("update %s cert file success", res.CertName)
	return nil
}

func copyCertToDockerDir(res util.ClientCertResp) error {
	hwlog.RunLog.Info("start to copy cert to docker's certs path")
	checker := fileutils.NewFileLinkChecker(false)
	checker.SetNext(fileutils.NewFileModeChecker(true, constants.ModeUmask022, true, true))
	checker.SetNext(fileutils.NewFileOwnerChecker(true, false, constants.RootUserUid, constants.RootUserGid))

	dockerCertPath := filepath.Join(constants.DockerCertDir, res.ImageAddress, constants.ImageCertFileName)
	if err := fileutils.MakeSureDir(dockerCertPath, checker); err != nil {
		return fmt.Errorf("create docker certs dir failed, error: %v", err)
	}
	certDir, err := path.GetCompSpecificDir(constants.ConfigCertPathName)
	if err != nil {
		return fmt.Errorf("get docker certs dir failed, error: %v", err)
	}
	caFilePath := filepath.Join(certDir, res.CertName, constants.RootCaName)
	if err = fileutils.CopyFile(caFilePath, dockerCertPath, checker); err != nil {
		return fmt.Errorf("copy cert [%s] to docker cert [%s] failed, error: %v", caFilePath, dockerCertPath, err)
	}
	if err = fileutils.SetPathPermission(dockerCertPath, constants.Mode400, false, false); err != nil {
		return fmt.Errorf("set path [%s] permission failed, error: %v", dockerCertPath, err)

	}
	hwlog.RunLog.Info("copy cert to docker's certs path success")
	return nil
}

func checkCrl(res util.ClientCertResp) error {
	if len(res.CrlContent) == 0 {
		return errors.New("empty crl content")
	}
	if len(res.CrlContent) > constants.CrlSizeLimit {
		hwlog.RunLog.Errorf("crl content size [%v bytes] exceeds limit", len(res.CrlContent))
		return errors.New("crl content size exceeds limit")
	}
	crlMgr, err := x509.NewCrlMgr([]byte(res.CrlContent))
	if err != nil {
		hwlog.RunLog.Errorf("parse crl content failed: %v", err)
		return errors.New("parse crl content failed")
	}
	return crlMgr.CheckCrl(x509.CertData{CertContent: []byte(res.CertContent)})
}

func saveCrl(res util.ClientCertResp) error {
	certDir, err := path.GetCompSpecificDir(constants.ConfigCertPathName)
	if err != nil {
		hwlog.RunLog.Errorf("get root-ca cert dir failed: %v", err)
		return errors.New("get root-ca cert dir failed")
	}
	crlFilePath := filepath.Join(certDir, res.CertName, constants.CommonRevocationListName)
	if err := fileutils.MakeSureDir(crlFilePath); err != nil {
		hwlog.RunLog.Errorf("create %s ca crl folder failed, error: %v", filepath.Base(crlFilePath), err)
		return errors.New("create ca crl folder failed")
	}

	if err := fileutils.WriteData(crlFilePath, []byte(res.CrlContent)); err != nil {
		hwlog.RunLog.Errorf("update %s crl file failed, error: %v", res.CertName, err)
		return errors.New("update crl file failed")
	}

	if err := backuputils.BackUpFiles(crlFilePath); err != nil {
		hwlog.RunLog.Warnf("create backup for %s crl failed, error: %v", res.CertName, err)
	}

	hwlog.RunLog.Infof("update %s crl file success", res.CertName)
	return nil
}

// Parse [method] implement of interface, parse message
func (ch *saveCertHandlerSdk) Parse(req *model.Message) error {
	hwlog.RunLog.Info("start to parse cert")
	ch.res = util.ClientCertResp{}
	if err := req.ParseContent(&ch.res); err != nil {
		hwlog.RunLog.Errorf("convert request param failed, error: %v", err)
		return err
	}
	return nil
}

// Check [method] implement of interface, check message
func (ch *saveCertHandlerSdk) Check(*model.Message) error {
	hwlog.RunLog.Info("start to check cert")
	if checkResult := newCertSdkParaChecker().Check(ch.res); !checkResult.Result {
		hwlog.RunLog.Errorf("check cert info failed: %v", checkResult.Reason)
		return errors.New("check cert info failed")
	}
	return nil
}

// PrintOpLogOk [method] implement for opLog success
func (ch *saveCertHandlerSdk) PrintOpLogOk() {
	hwlog.OpLog.Infof("[%s@%s] %s %s cert success",
		config.NetMgr.NetType, config.NetMgr.IP, ch.res.CertOpt, ch.res.CertName)
}

// PrintOpLogFail [method] implement for opLog failed
func (ch *saveCertHandlerSdk) PrintOpLogFail() {
	hwlog.OpLog.Errorf("[%s@%s] %s %s cert failed",
		config.NetMgr.NetType, config.NetMgr.IP, ch.res.CertOpt, ch.res.CertName)
}
