// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package x509 for certificate utils
package x509

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strings"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/utils"
)

// used to check service cert
const (
	v3 = 3
)

// CheckSvcCertTask to check service cert
type CheckSvcCertTask struct {
	KeyPath              string
	SvcCertData          []byte
	KmcConfig            *kmc.SubConfig
	svcCert              *x509.Certificate
	AllowFutureEffective bool
}

// RunTask to check service cert step by step
func (cct *CheckSvcCertTask) RunTask() error {
	if err := cct.initTask(); err != nil {
		return err
	}

	var checkFunc = []func() error{
		cct.checkExtension,
		cct.verifyCertWithKey,
		cct.checkOverdueTime,
		cct.checkSignAlgoAndKeyLength,
	}

	for _, function := range checkFunc {
		if err := function(); err != nil {
			return err
		}
	}
	return nil
}
func (cct *CheckSvcCertTask) checkExtension() error {
	if cct.svcCert.BasicConstraintsValid && cct.svcCert.IsCA {
		return errors.New("server cert cannot be ca file")
	}

	if cct.svcCert.Version != v3 {
		return errors.New("the cert must be v3")
	}
	if (cct.svcCert.KeyUsage&x509.KeyUsageCertSign) == x509.KeyUsageCertSign ||
		(cct.svcCert.KeyUsage&x509.KeyUsageCRLSign) == x509.KeyUsageCRLSign {
		return errors.New("the cert keyUsage is invalid")
	}

	hwlog.RunLog.Info("check extension success")
	return nil
}

func (cct *CheckSvcCertTask) initTask() error {
	var err error

	block, _ := pem.Decode(cct.SvcCertData)
	if block == nil {
		hwlog.RunLog.Errorf("service cert content is empty")
		return errors.New("service cert content is empty")
	}
	cct.svcCert, err = x509.ParseCertificate(block.Bytes)

	if err != nil {
		hwlog.RunLog.Errorf("parse certificate failed: %v", err)
		return errors.New("parse certificate failed")
	}
	return nil
}

func (cct *CheckSvcCertTask) verifyCertWithKey() error {
	encryptedKeyByte, err := fileutils.LoadFile(cct.KeyPath)
	if err != nil {
		return errors.New("load key file to byte failed")
	}

	decryptedKeyByte, err := kmc.DecryptContent(encryptedKeyByte, cct.KmcConfig)
	if err != nil {
		return errors.New("decrypt key content failed")
	}
	defer utils.ClearSliceByteMemory(decryptedKeyByte)

	_, err = tls.X509KeyPair(cct.SvcCertData, decryptedKeyByte)
	if err != nil {
		return errors.New("the key and the cert do not match")
	}

	hwlog.RunLog.Info("verify cert with key success")
	return nil
}

func (cct *CheckSvcCertTask) checkOverdueTime() error {
	if err := CheckValidityPeriod(cct.svcCert, cct.AllowFutureEffective); err != nil {
		hwlog.RunLog.Errorf("check service cert overdue time failed: %v", err)
		return errors.New("check service cert overdue time failed")
	}

	hwlog.RunLog.Info("check overdue time success")
	return nil
}

func (cct *CheckSvcCertTask) checkSignAlgoAndKeyLength() error {
	signAlg := cct.svcCert.SignatureAlgorithm.String()
	if signAlg == "0" {
		hwlog.RunLog.Error("the hash algorithm is not support")
		return errors.New("the hash algorithm is not support")
	}

	insecureHashAlgos := []string{"SHA0", "SHA1", "MD2", "MD4", "MD5", "RIPEMD", "RIPEMD-128"}
	for _, insecureHashAlgo := range insecureHashAlgos {
		if strings.Contains(signAlg, insecureHashAlgo) {
			hwlog.RunLog.Error("the hash algorithm is insecure")
			return errors.New("the hash algorithm is insecure")
		}
	}
	if err := CheckPubKeyLength(cct.svcCert); err != nil {
		hwlog.RunLog.Error(err)
		return err
	}

	hwlog.RunLog.Info("check signature algorithm and key length success")
	return nil
}
