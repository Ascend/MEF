// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tls provides the capability of tls.
package tls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"

	"huawei.com/mindx/common/fileutils"
	cx509 "huawei.com/mindx/common/x509"
)

const (
	rsaLength = 3072
	eccLength = 256
)

// LoadCertPair load and valid encrypted certificate and private key
// parameter pathMap key is cx509.CertStorePath,cx509.CertStoreBackupPath,cx509.KeyStorePath,cx509.KeyStoreBackupPath
// cx509.PassFilePath,cx509.PassFileBackUpPath
func LoadCertPair(pathMap map[string]string, encryptAlgorithm int) (*tls.Certificate, error) {
	certBytes, keyPem, err := cx509.LoadCertPairByte(pathMap, encryptAlgorithm, fileutils.Mode600)
	if err != nil {
		return nil, err
	}
	defer cx509.PaddingAndCleanSlice(keyPem)
	return ValidateCertPair(certBytes, keyPem, true, cx509.InvalidNum)
}

// ValidateCertPair Validate the cert pair
func ValidateCertPair(certBytes, keyPem []byte, periodCheck bool, overdueTime int) (*tls.Certificate, error) {
	var err error
	var tlsCert *tls.Certificate
	// preload cert and key files
	tlsCert, err = ValidateX509Pair(certBytes, keyPem, overdueTime)
	if err != nil || tlsCert == nil {
		return nil, err
	}
	x509Cert, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		return nil, errors.New("parse certificate failed")
	}
	if err = cx509.AddToCertStatusTrace(x509Cert); err != nil {
		return nil, err
	}
	if periodCheck {
		go cx509.PeriodCheck()
	}
	return tlsCert, nil
}

// ValidateX509Pair validate the x509pair
func ValidateX509Pair(certBytes []byte, keyBytes []byte, overdueTime int) (*tls.Certificate, error) {
	c, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		return nil, errors.New("failed to load X509KeyPair")
	}
	cc, err := x509.ParseCertificate(c.Certificate[0])
	if err != nil {
		return nil, errors.New("parse certificate failed")
	}
	if err = cx509.CheckExtension(cc); err != nil {
		return nil, err
	}
	if err = cx509.CheckSignatureAlgorithm(cc); err != nil {
		return nil, err
	}
	switch overdueTime {
	case cx509.InvalidNum:
		err = cx509.CheckValidityPeriod(cc, false)
	default:
		err = cx509.CheckValidityPeriodWithError(cc, overdueTime)
	}
	if err != nil {
		return nil, err
	}
	keyLen, keyType, err := cx509.GetPrivateKeyLength(cc, &c)
	if err != nil {
		return nil, err
	}
	// ED25519 private key length is stable and no need to verify
	if "RSA" == keyType && keyLen < rsaLength || "ECC" == keyType && keyLen < eccLength {
		return nil, errors.New("the private key length is not enough")
	}
	return &c, nil
}
