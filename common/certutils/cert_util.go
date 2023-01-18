// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package certutils Cert utils
package certutils

import (
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"

	"huawei.com/mindx/common/rand"
	"huawei.com/mindx/common/utils"
	hwCertMgr "huawei.com/mindx/common/x509"

	"huawei.com/mindxedge/base/common"
)

func getCsr(priv *rsa.PrivateKey, commonName string, dnsName string) ([]byte, error) {
	template := x509.CertificateRequest{
		Subject:  pkix.Name{CommonName: commonName},
		DNSNames: []string{dnsName},
	}
	csrDer, err := x509.CreateCertificateRequest(rand.Reader, &template, priv)
	if err != nil {
		return nil, errors.New("generate csr for self signed certificate failed: " + err.Error())
	}
	return csrDer, nil
}

// CreateCsr [method] for create csr content
func CreateCsr(keyPath string, commonName string, dnsName string, kmcCfg *common.KmcCfg) ([]byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, priKeyLength)
	if err != nil {
		return nil, errors.New("generate new key for self signed certificate failed: " + err.Error())
	}

	csr, err := getCsr(priv, commonName, dnsName)
	if err != nil {
		return nil, err
	}

	err = saveKeyWithPem(keyPath, priv, kmcCfg)
	if err != nil {
		return nil, errors.New("save self singed key with pem failed: " + err.Error())
	}

	return csr, nil
}

// GetCertPair get cert and key pair info
func GetCertPair(certPath, keyPath string, kmcCfg *common.KmcCfg) (*CaPairInfo, error) {
	pairWithPem, err := GetCertPairForPem(certPath, keyPath, kmcCfg)
	if err != nil {
		return nil, err
	}

	caCert, err := hwCertMgr.LoadCertsFromPEM(pairWithPem.CertPem)
	if err != nil {
		return nil, errors.New("decode cert form pem failed: " + err.Error())
	}

	caPrivate := PemUnwrapPrivKey(pairWithPem.KeyPem)
	if caPrivate == nil {
		return nil, errors.New("unwrap a private key pem failed")
	}
	var caPairInfo = &CaPairInfo{
		Cert:   caCert,
		PriKey: caPrivate,
	}
	return caPairInfo, nil
}

// GetCertPairForPem [method] for get cert pair
func GetCertPairForPem(certPath, keyPath string, kmcCfg *common.KmcCfg) (*CaPairInfoWithPem, error) {
	certBytes, err := utils.LoadFile(certPath)
	if certBytes == nil {
		return nil, fmt.Errorf("load cert path [%s] file failed", certPath)
	}
	encryptKeyContent, err := utils.LoadFile(keyPath)
	if encryptKeyContent == nil {
		return nil, fmt.Errorf("load key path [%s] file failed: ", keyPath)
	}
	decryptKeyByte, err := common.DecryptContent(encryptKeyContent, kmcCfg)
	if err != nil {
		return nil, errors.New("decrypt key content failed: " + err.Error())
	}
	return &CaPairInfoWithPem{CertPem: certBytes, KeyPem: decryptKeyByte}, nil
}

// PemWrapCert code der to pem type
func PemWrapCert(der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  pubCertType,
		Bytes: der,
	})
}

// PemWrapPrivKey code der private key to pem type
func PemWrapPrivKey(priv *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  privKeyType,
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})
}

// PemUnwrapPrivKey decode pem private key to der type
func PemUnwrapPrivKey(p []byte) *rsa.PrivateKey {
	pm, _ := pem.Decode(p)
	if pm == nil {
		return nil
	}

	if pm.Type != privKeyType {
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
	err = utils.MakeSureDir(keyPath)
	if err != nil {
		return err
	}
	err = hwCertMgr.OverridePassWdFile(keyPath, encryptKeyPem, fileMode)
	if err != nil {
		return err
	}
	return nil
}
