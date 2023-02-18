// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package certmanager cert manager module
package certmanager

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509"

	"cert-manager/pkg/certconstant"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/certutils"
	"huawei.com/mindxedge/base/common/httpsmgr"
)

// getCertByCertName query root ca with use id
func getCertByCertName(certName string) ([]byte, error) {
	if !CheckCertName(certName) {
		hwlog.RunLog.Error("the cert name not support")
		return nil, errors.New("the cert name not support")
	}
	caFilePath := getRootCaPath(certName)
	certData, err := utils.LoadFile(caFilePath)
	if certData == nil {
		hwlog.RunLog.Errorf("load root cert failed: %v", err)
		return nil, errors.New("load root cert failed")
	}
	return certData, nil
}

// issueServiceCert issue service certificate with csr file, only support pem type csr
func issueServiceCert(certName string, serviceCsr string) ([]byte, error) {
	if !CheckCertName(certName) {
		hwlog.RunLog.Errorf("issue service cert failed, check cert name [%s] failed", certName)
		return nil, errors.New("check cert name failed, not a support cert name")
	}

	srvDer, _ := pem.Decode([]byte(serviceCsr))
	if srvDer == nil {
		hwlog.RunLog.Error("Decode service csr pem failed")
		return nil, errors.New("decode service csr pem failed")
	}

	keyFilePath := getRootKeyPath(certName)
	caFilePath := getRootCaPath(certName)
	initCertMgr := certutils.InitRootCertMgr(caFilePath, keyFilePath, certName, nil)
	if _, err := initCertMgr.GetRootCaPair(); err != nil {
		if _, err = initCertMgr.NewRootCa(); err != nil {
			hwlog.RunLog.Errorf("Init root ca info failed: %v", err)
			return nil, err
		}
	}

	certBytes, err := initCertMgr.IssueServiceCert(srvDer.Bytes)
	if err != nil {
		hwlog.RunLog.Errorf("issue service cert info failed: %v", err)
		return nil, err
	}

	certPem := certutils.PemWrapCert(certBytes)
	hwlog.RunLog.Info("issue service cert success")
	return certPem, nil
}

// CheckAndCreateRootCa [method] for check root ca
func CheckAndCreateRootCa() error {
	hwlog.RunLog.Infof("start to check cert: %s", common.WsCltName)
	err := checkAndCreateCa(common.WsCltName)
	if err != nil {
		hwlog.RunLog.Errorf("check cert [%s] failed: %v", common.WsCltName, err)
		return err
	}
	hwlog.RunLog.Infof("check cert [%s] success", common.WsCltName)
	return nil
}

func checkAndCreateCa(certName string) error {
	var certPath, keyPath string
	if certName == common.InnerName {
		certPath = getInnerRootCaPath()
	} else {
		certPath = getRootCaPath(certName)
	}
	_, err := utils.CheckPath(certPath)
	if err != nil {
		return err
	}
	innerCert, err := utils.LoadFile(certPath)
	if innerCert != nil {
		// todo 增加备份恢复的校验
		return nil
	}
	//  证书不存在，生成根证书
	if certName == common.InnerName {
		keyPath = getInnerRootKeyPath()
	} else {
		keyPath = getRootKeyPath(certName)
	}
	initCertMgr := certutils.InitRootCertMgr(certPath, keyPath, certName, nil)
	_, err = initCertMgr.NewRootCa()
	if err != nil {
		return err
	}
	return nil
}

// saveCaContent save ca content to File
func saveCaContent(certName string, caContent []byte) error {
	caFilePath := path.Join(certconstant.RootCaMgrDir, certName, certconstant.RootCaFileName)
	if err := utils.MakeSureDir(caFilePath); err != nil {
		hwlog.RunLog.Errorf("create %s ca folder failed, error: %v", path.Base(caFilePath), err)
		return fmt.Errorf("create %s ca folder failed, error: %v", path.Base(caFilePath), err)
	}
	caFile, err := os.OpenFile(caFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, common.Mode600)
	if err != nil {
		hwlog.RunLog.Errorf("create %s ca file failed", path.Base(caFilePath))
		return fmt.Errorf("create %s ca file failed", path.Base(caFilePath))
	}
	defer func(caFile *os.File) {
		if err := caFile.Close(); err != nil {
			hwlog.RunLog.Errorf("caFile close failed, error: %v", err)
		}
	}(caFile)
	caWriter := bufio.NewWriter(caFile)
	if _, err := io.Copy(caWriter, bytes.NewReader(caContent)); err != nil {
		hwlog.RunLog.Errorf("write ca file failed, path: %s, error: %v", path.Base(caFilePath), err)
		return fmt.Errorf("write ca file failed, path: %s, error: %v", path.Base(caFilePath), err)
	}
	if err := caWriter.Flush(); err != nil {
		hwlog.RunLog.Errorf("flush %s ca file failed, error: %v", path.Base(caFilePath), err)
		return fmt.Errorf("flush %s ca file failed, error: %v", path.Base(caFilePath), err)
	}
	hwlog.RunLog.Infof("finished ca content copy to %s", caFilePath)
	return nil
}

// checkCert check cert content
func checkCert(req importCertReq) ([]byte, error) {
	// verifying the certificate usage type
	if !CheckCertName(req.CertName) {
		hwlog.RunLog.Error("valid cert name parameter failed")
		return []byte{}, errors.New("valid cert name parameter failed")
	}
	// base64 decode root certificate content
	caBase64, err := base64.StdEncoding.DecodeString(req.Cert)
	if err != nil {
		hwlog.RunLog.Errorf("base64 decode %s ca content failed, err:%v", err)
		return []byte{}, errors.New("base64 decode ca content failed")
	}
	if len(caBase64) == 0 || len(caBase64) > certutils.CertSizeLimited {
		hwlog.RunLog.Errorf("valid ca file size failed")
		return []byte{}, errors.New("valid ca file size failed")
	}
	// verifying root certificate content
	if err := x509.VerifyCaCert(caBase64, x509.InvalidNum); err != nil {
		hwlog.RunLog.Errorf("valid ca certification failed, error:%v", err)
		return []byte{}, errors.New("valid ca certification failed")
	}
	hwlog.RunLog.Info("valid cert success")
	return caBase64, nil
}

// getCertContent query cert content with cert name
func getCertContent(certName string) (string, error) {
	if !CheckCertName(certName) {
		hwlog.RunLog.Error("the cert name not support")
		return "", errors.New("the cert name not support")
	}
	caFilePath := getRootCaPath(certName)
	certData, err := utils.LoadFile(caFilePath)
	if certData == nil {
		hwlog.RunLog.Errorf("load root cert failed: %v", err)
		return "", errors.New("load root cert failed")
	}
	return string(certData), nil
}

// removeCaFile delete ca File
func removeCaFile(certName string) error {
	caFilePath := getRootCaPath(certName)
	if utils.IsExist(caFilePath) {
		// remove the ca file
		if err := common.DeleteFile(caFilePath); err != nil {
			hwlog.RunLog.Errorf("remove %s ca file failed, error: %v", path.Base(caFilePath), err)
			return fmt.Errorf("remove %s ca file failed, error: %v", path.Base(caFilePath), err)
		}
	}
	hwlog.RunLog.Infof("delete %s ca file success", caFilePath)
	return nil
}

func updateClientCert(certName string, certContent []byte) error {
	hwlog.RunLog.Info("start update cert file")
	reqCertParams := httpsmgr.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath:    certconstant.RootCaPath,
			CertPath:      certconstant.ServerCertPath,
			KeyPath:       certconstant.ServerKeyPath,
			SvrFlag:       false,
			IgnoreCltCert: false,
			KmcCfg:        nil,
		},
	}
	cert := certutils.UpdateClientCert{
		CertName:    certName,
		CertContent: certContent,
	}
	for i := 0; i < certutils.DefaultCertRetryTime; i++ {
		_, err := reqCertParams.UpdateCertFile(cert)
		if err == nil {
			break
		}
		time.Sleep(certutils.DefaultCertWaitTime)
	}
	return nil
}
