// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package certmanager cert manager module
package certmanager

import (
	"encoding/pem"
	"errors"
	"fmt"
	"path"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"cert-manager/pkg/certconstant"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/certutils"
	"huawei.com/mindxedge/base/common/httpsmgr"
)

// getCertByCertName query root ca with use id
func getCertByCertName(certName string) ([]byte, error) {
	caFilePath := getRootCaPath(certName)
	if (certName == common.ImageCertName || certName == common.SoftwareCertName) && !utils.IsExist(caFilePath) {
		hwlog.RunLog.Warnf("%s cert content should be imported", certName)
		return nil, nil
	}
	certData, err := utils.LoadFile(caFilePath)
	if certData == nil {
		hwlog.RunLog.Errorf("load root cert failed: %v", err)
		return nil, errors.New("load root cert failed")
	}
	return certData, nil
}

// issueServiceCert issue service certificate with csr file, only support pem type csr
func issueServiceCert(certName string, serviceCsr string) ([]byte, error) {
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
	caFilePath := getRootCaPath(certName)
	if err := utils.MakeSureDir(caFilePath); err != nil {
		hwlog.RunLog.Errorf("create %s ca folder failed, error: %v", path.Base(caFilePath), err)
		return fmt.Errorf("create %s ca folder failed, error: %v", path.Base(caFilePath), err)
	}
	if err := common.WriteData(caFilePath, caContent); err != nil {
		hwlog.RunLog.Errorf("save %s cert file failed, error:%s", certName, err)
		return fmt.Errorf("save %s ca file failed", certName)
	}
	hwlog.RunLog.Infof("save %s ca file success", certName)
	return nil
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

func updateClientCert(certName, certOpt string, certContent []byte) error {
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
		CertOpt:     certOpt,
	}
	certName, err := reqCertParams.UpdateCertFile(cert)
	if err != nil {
		hwlog.RunLog.Errorf("update %s ca file failed, error:%v", certName, err)
		return fmt.Errorf("update %s ca file failed", certName)
	}
	return nil
}
