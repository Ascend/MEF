// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package certmgr Package cert_mgr cert manager module
package certmgr

import (
	"encoding/pem"
	"errors"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/certutils"
)

// QueryRootCa query root ca with use id
func QueryRootCa(certName string) ([]byte, error) {
	if !checkCertName(certName) {
		hwlog.RunLog.Error("check cert name failed, the cert name not support")
		return nil, errors.New("query root ca failed: the cert name not support")
	}
	caFilePath := getRootCaPath(certName)
	certData, err := utils.LoadFile(caFilePath)
	if certData == nil {
		hwlog.RunLog.Errorf("load root cert failed: %v", err)
		return nil, errors.New("load root cert failed")
	}
	return certData, nil
}

// IssueServiceCert issue service certificate with csr file(only support pem type csr)
func IssueServiceCert(certName string, serviceCsr string) ([]byte, error) {
	if !checkCertName(certName) {
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
	for certName := range certImportMap {
		hwlog.RunLog.Infof("start to check cert: %s", certName)
		err := checkAndCreateCa(certName)
		if err != nil {
			hwlog.RunLog.Errorf("check cert [%s] failed: %v", certName, err)
			return err
		}
		hwlog.RunLog.Infof("check cert [%s] success", certName)
	}
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
