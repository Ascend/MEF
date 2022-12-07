// Copyright (c) 2021. Huawei Technologies Co., Ltd. All rights reserved.

// Package certutils Cert utils
package certutils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	hwCertMgr "huawei.com/mindx/common/x509"
	"huawei.com/mindxedge/base/common"
)

const (
	// PriKeyLength private key length
	PriKeyLength = 4096
	// ValidationYearCA root ca validate year
	ValidationYearCA = 10
	// ValidationYearCert service Cert validate year
	ValidationYearCert = 10
	// ValidationMonth Cert validate month
	ValidationMonth = 0
	// ValidationDay Cert validate day
	ValidationDay = 0
	// BigIntSize server_number
	BigIntSize = 2022
	// CaCountry issue country
	CaCountry = "CN"
	// CaOrganization issue organization
	CaOrganization = "Huawei"
	// CaOrganizationalUnit issue unit
	CaOrganizationalUnit = "Ascend"
	// CaCommonName issue name
	CaCommonName = "MEF"
	// PubCertType Cert type
	PubCertType = "CERTIFICATE"
	// PrivKeyType Cert key type
	PrivKeyType = "RSA PRIVATE KEY"
	// FileMode Cert file mode
	FileMode = 0600
)

// PemWrapCert code der to pem type
func PemWrapCert(der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  PubCertType,
		Bytes: der,
	})
}

// PemUnwrapCert decode pem to der type
func PemUnwrapCert(p []byte) ([]byte, []byte) {
	pm, r := pem.Decode(p)
	if pm == nil {
		return nil, r
	}

	if pm.Type != PubCertType {
		return nil, r
	}

	return pm.Bytes, r
}

// PemWrapPrivKey code der private key to pem type
func PemWrapPrivKey(priv *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  PrivKeyType,
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})
}

// PemUnwrapPrivKey decode pem private key to der type
func PemUnwrapPrivKey(p []byte) *rsa.PrivateKey {
	pm, _ := pem.Decode(p)
	if pm == nil {
		return nil
	}

	if pm.Type != PrivKeyType {
		return nil
	}

	privKey, err := x509.ParsePKCS1PrivateKey(pm.Bytes)
	if err != nil {
		return nil
	}

	return privKey
}

func saveCertWithPem(certPath string, certDerBytes []byte) error {
	certPem := PemWrapCert(certDerBytes)
	err := common.WriteData(certPath, certPem)
	if err != nil {
		return err
	}
	return nil
}

func saveKeyWithPem(keyPath string, keyDerBytes *rsa.PrivateKey, kmcCfg *common.KmcCfg) error {
	keyPem := PemWrapPrivKey(keyDerBytes)
	defer hwCertMgr.PaddingAndCleanSlice(keyPem)
	encryptKeyPem, err := common.EncryptContent(keyPem, kmcCfg)
	if err != nil {
		return err
	}
	err = hwCertMgr.OverridePassWdFile(keyPath, encryptKeyPem, FileMode)
	if err != nil {
		return err
	}
	return nil
}
