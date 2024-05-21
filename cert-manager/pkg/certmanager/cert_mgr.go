// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package certmanager cert manager module
package certmanager

import (
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509/certutils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"

	"cert-manager/pkg/certmanager/certchecker"
)

var (
	lock            sync.Mutex
	certsToImported = map[string]struct{}{
		common.ImageCertName:    {},
		common.SoftwareCertName: {},
		common.NorthernCertName: {},
	}
)

func isCertImported(certName string) bool {
	caFilePath := getRootCaPath(certName)
	_, exist := certsToImported[certName]
	return !exist || fileutils.IsExist(caFilePath) || fileutils.IsExist(caFilePath+backuputils.BackupSuffix)
}

// isExternalCrlImported checks whether an external crl is available.
func isExternalCrlImported(crlName string) bool {
	crlFilePath := getCrlPath(crlName)
	_, exist := certsToImported[crlName]
	return exist && (fileutils.IsExist(crlFilePath) || fileutils.IsExist(crlFilePath+backuputils.BackupSuffix))
}

// getCertByCertName query root ca with cert name
func getCertByCertName(certName string) ([]byte, error) {
	caFilePath := getRootCaPath(certName)
	certData, err := certutils.GetCertContentWithBackup(caFilePath)
	if err != nil {
		hwlog.RunLog.Errorf("load root cert failed: %v", err)
		return nil, fmt.Errorf("load root cert failed: %v", err)
	}
	return certData, nil
}

// CreateCaIfNotExit create new ca with specified name
func CreateCaIfNotExit(certName string) error {
	keyFilePath := getRootKeyPath(certName)
	certFilePath := getRootCaPath(certName)
	if isRootCaFilesExist(certFilePath, keyFilePath) {
		return nil
	}
	initCertMgr := certutils.InitRootCertMgr(certFilePath, keyFilePath, common.MefCertCommonNamePrefix, nil)
	_, err := initCertMgr.NewRootCaWithBackup()
	return err
}

// CreateTempCaCert create temp ca certs with .tmp suffix, return cert bytes
func CreateTempCaCert(caCertName string) (string, error) {
	tempKeyPath := getTempRootKeyPath(caCertName)
	tempCertPath := getTempRootCaPath(caCertName)
	// if previously created certs exist, use them, otherwise create new certs.
	if fileutils.IsExist(tempKeyPath) && fileutils.IsExist(tempCertPath) {
		certData, err := fileutils.LoadFile(tempCertPath)
		if err != nil {
			return "", fmt.Errorf("load previously created temp root ca failed: %v", err)
		}
		return string(certData), nil
	}

	if _, err := certutils.InitRootCertMgr(tempCertPath, tempKeyPath, caCertName, nil).NewRootCa(); err != nil {
		return "", fmt.Errorf("create new root ca [%v] failed: %v", caCertName, err)
	}
	certData, err := fileutils.LoadFile(tempCertPath)
	if err != nil {
		return "", fmt.Errorf("load new temp root ca failed: %v", err)
	}
	return string(certData), nil
}

// UpdateCaCertWithTemp replace old certs with temporary new certs
func UpdateCaCertWithTemp(certName string) error {
	oldKeyFilePath := getRootKeyPath(certName)
	oldCertFilePath := getRootCaPath(certName)
	tempKeyFilePath := getTempRootKeyPath(certName)
	tempCertFilePath := getTempRootCaPath(certName)
	if err := fileutils.SetPathPermission(oldCertFilePath, common.Mode600, false, false); err != nil {
		return err
	}
	if err := fileutils.SetPathPermission(oldKeyFilePath, common.Mode600, false, false); err != nil {
		return err
	}
	defer func() {
		if err := fileutils.SetPathPermission(oldCertFilePath, common.Mode400, false, false); err != nil {
			hwlog.RunLog.Errorf("set cert file [%v] permission to 400 error: %v", oldCertFilePath, err)
		}
		if err := fileutils.SetPathPermission(oldKeyFilePath, common.Mode400, false, false); err != nil {
			hwlog.RunLog.Errorf("set key file [%v] permission to 400 error: %v", oldKeyFilePath, err)
		}
		if err := RemoveTempCaCert(certName); err != nil {
			hwlog.RunLog.Errorf("delete temp ca cert error: %v", err)
		}
	}()
	if err := fileutils.CopyFile(tempKeyFilePath, oldKeyFilePath); err != nil {
		return err
	}
	if err := fileutils.CopyFile(tempCertFilePath, oldCertFilePath); err != nil {
		return err
	}
	if err := backuputils.BackUpFiles(oldCertFilePath, oldKeyFilePath); err != nil {
		hwlog.RunLog.Warnf("create backup files for upgraded cert failed, %v", err)
	}

	return nil
}

// RemoveTempCaCert delete all temporary ca certs
func RemoveTempCaCert(certName string) error {
	tempKeyFilePath := getTempRootKeyPath(certName)
	tempCertFilePath := getTempRootCaPath(certName)
	if fileutils.IsExist(tempKeyFilePath) {
		if err := fileutils.DeleteAllFileWithConfusion(tempKeyFilePath); err != nil {
			return err
		}
	}
	if fileutils.IsExist(tempCertFilePath) {
		if err := fileutils.DeleteFile(tempCertFilePath); err != nil {
			return err
		}
	}
	return nil
}

// issueServiceCert issue service certificate with csr file, only support pem type csr
func issueServiceCert(certName string, serviceCsr string) ([]byte, error) {
	csrByte, err := base64.StdEncoding.DecodeString(serviceCsr)
	if err != nil {
		hwlog.RunLog.Error("base64 decode service csr failed")
		return nil, errors.New("base64 decode service csr failed")
	}
	srvDer, _ := pem.Decode(csrByte)
	if srvDer == nil {
		hwlog.RunLog.Error("Decode service csr pem failed")
		return nil, errors.New("decode service csr pem failed")
	}

	keyFilePath := getRootKeyPath(certName)
	caFilePath := getRootCaPath(certName)
	// use hub_client new root ca to issue service certs when cert update is in process
	if certName == common.WsCltName || certName == common.WsSerName {
		tempKeyPath := getTempRootKeyPath(certName)
		tempCertPath := getTempRootCaPath(certName)
		if fileutils.IsExist(tempKeyPath) && fileutils.IsExist(tempCertPath) {
			keyFilePath = tempKeyPath
			caFilePath = tempCertPath
		}
	}
	initCertMgr := certutils.InitRootCertMgr(caFilePath, keyFilePath, common.MefCertCommonNamePrefix, nil)
	if !isRootCaFilesExist(caFilePath, keyFilePath) {
		if _, err = initCertMgr.NewRootCaWithBackup(); err != nil {
			hwlog.RunLog.Errorf("Init root ca info failed: %v", err)
			return nil, err
		}
	}

	certBytes, err := initCertMgr.IssueServiceCertWithBackup(srvDer.Bytes)
	if err != nil {
		hwlog.RunLog.Errorf("issue service cert info failed: %v", err)
		return nil, err
	}

	certPem := certutils.PemWrapCert(certBytes)
	hwlog.RunLog.Info("issue service cert success")
	return certPem, nil
}

func isRootCaFilesExist(caFilePath, keyFilePath string) bool {
	paths := []string{caFilePath, caFilePath + backuputils.BackupSuffix,
		keyFilePath, keyFilePath + backuputils.BackupSuffix}
	for _, path := range paths {
		if fileutils.IsLexist(path) {
			return true
		}
	}
	return false
}

// saveCaContent save ca content to File
func saveCaContent(certName string, caContent []byte) error {
	caFilePath := getRootCaPath(certName)
	if err := fileutils.MakeSureDir(caFilePath); err != nil {
		hwlog.RunLog.Errorf("create %s ca folder failed, error: %v", certName, err)
		return fmt.Errorf("create %s ca folder failed, error: %v", certName, err)
	}
	if err := fileutils.WriteData(caFilePath, caContent); err != nil {
		hwlog.RunLog.Errorf("save %s cert file failed, error:%s", certName, err)
		return fmt.Errorf("save %s ca file failed", certName)
	}
	if err := backuputils.BackUpFiles(caFilePath); err != nil {
		hwlog.RunLog.Warnf("back up %s ca file failed, %v", caFilePath, err)
	}
	hwlog.RunLog.Infof("save %s ca file success", certName)
	return nil
}

// removeCaFile delete ca File
func removeCaFile(certName string) error {
	caFilePath := getRootCaPath(certName)
	toDeleteFiles := []string{caFilePath, caFilePath + backuputils.BackupSuffix}
	for _, filePath := range toDeleteFiles {
		if err := fileutils.DeleteFile(filePath); err != nil {
			hwlog.RunLog.Errorf("remove %s ca file failed, error: %v", certName, err)
			return fmt.Errorf("remove %s ca file failed, error: %v", certName, err)
		}
	}
	hwlog.RunLog.Infof("delete %s ca file success", caFilePath)
	return nil
}

func updateClientCert(certName, operation string) error {
	lock.Lock()
	defer lock.Unlock()
	hwlog.RunLog.Info("start update cert file")
	reqCertParams := requests.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath: util.RootCaPath,
			CertPath:   util.ServerCertPath,
			KeyPath:    util.ServerKeyPath,
			SvrFlag:    false,
			KmcCfg:     nil,
		},
	}
	updateClientCertReq := certutils.UpdateClientCert{CertName: certName, CertOpt: operation}
	if operation == common.Update {
		certContent, err := getCertByCertName(certName)
		if err != nil {
			hwlog.RunLog.Errorf("load %s ca file failed, error:%v", certName, err)
			return fmt.Errorf("load %s ca file failed", certName)
		}
		updateClientCertReq.CertContent = certContent

		// CrlContent is optional field
		if isExternalCrlImported(certName) {
			crlContent, err := certutils.GetCrlContentWithBackup(getCrlPath(certName))
			if err != nil {
				hwlog.RunLog.Errorf("load %s crl file failed, error:%v", certName, err)
				return fmt.Errorf("load %s crl file failed", certName)
			}

			updateClientCertReq.CrlContent = crlContent
		}

	}

	certName, err := reqCertParams.UpdateCertFile(updateClientCertReq)
	if err != nil {
		hwlog.RunLog.Errorf("update %s ca file failed, error:%v", certName, err)
		return fmt.Errorf("update %s ca file failed", certName)
	}
	return nil
}

// ExportRootCa export cert file
func ExportRootCa(c *gin.Context) {
	hwlog.RunLog.Info("export cert file start")
	certName := c.Query("certName")
	if !certchecker.CheckIfCanExport(certName) {
		msg := fmt.Sprintf("export cert [%s] root ca not support", certName)
		hwlog.RunLog.Error(msg)
		common.ConstructResp(c, common.ErrorExportRootCa, msg, nil)
		return
	}

	var caBytes []byte
	var err error
	tempCaFilePath := getTempRootCaPath(certName)
	if fileutils.IsExist(tempCaFilePath) {
		if caBytes, err = certutils.GetCertContent(tempCaFilePath); err != nil {
			hwlog.RunLog.Errorf("load new temp root cert failed: %v, old root cert will be used", err)
		}
	}
	// if new temp ca exists, but get an error when load it, use old ca data.
	if len(caBytes) == 0 {
		if caBytes, err = getCertByCertName(certName); err != nil {
			msg := fmt.Sprintf("get cert [%s] root ca failed", certName)
			hwlog.RunLog.Errorf("%s, error: %v", msg, err)
			common.ConstructResp(c, common.ErrorExportRootCa, msg, nil)
			return
		}
	}
	c.Writer.WriteHeader(http.StatusOK)
	c.Header(common.ContentType, "text/plain; charset=utf-8")
	c.Header(common.ContentDisposition, fmt.Sprintf("attachment; filename=%s", util.RootCaFileName))
	c.Writer.WriteHeaderNow()
	if _, err = c.Writer.Write(caBytes); err != nil {
		msg := fmt.Sprintf("export cert [%s] root ca failed", certName)
		hwlog.RunLog.Errorf("%s, error: %v", msg, err)
		common.ConstructResp(c, common.ErrorExportRootCa, msg, nil)
		return
	}
	c.Writer.Flush()
	hwlog.RunLog.Infof("export cert [%s] root ca success", certName)
}
