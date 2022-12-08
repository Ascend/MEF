// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package certmgr Package cert_mgr cert manager module
package certmgr

import (
	"encoding/pem"
	"errors"
	"path"

	"cert-manager/pkg/certconstant"
	"cert-manager/pkg/certid"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindxedge/base/common/certutils"
)

// QueryRootCa query root ca with use id
func QueryRootCa(useId string) ([]byte, error) {
	if !certid.CheckUseId(useId) {
		hwlog.RunLog.Errorf("check cert use id failed, use id not support")
		return nil, errors.New("query root ca failed: use id not support")
	}
	certName := certid.GetUseIdName(useId)
	caFilePath := path.Join(certconstant.RootCaPath, certName, certconstant.RootCaFileName)
	certData, err := utils.LoadFile(caFilePath)
	if certData == nil {
		hwlog.RunLog.Errorf("load root cert failed: %v", err)
		return nil, errors.New("load root cert failed")
	}
	return certData, nil
}

// IssueServiceCert issue service certificate with csr file(only support pem type csr)
func IssueServiceCert(useId string, serviceCsrPem []byte) ([]byte, error) {
	srvDer, _ := pem.Decode(serviceCsrPem)
	if srvDer == nil {
		hwlog.RunLog.Error("Decode service csr pem failed")
		return nil, errors.New("decode service csr pem failed")
	}
	certName := certid.GetUseIdName(useId)
	keyFilePath := path.Join(certconstant.RootCaPath, certName, certconstant.RootKeyFileName)
	caFilePath := path.Join(certconstant.RootCaPath, certName, certconstant.RootCaFileName)
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
