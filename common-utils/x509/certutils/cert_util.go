// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package certutils Cert utils
package certutils

import (
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/rand"
	"huawei.com/mindx/common/utils"
	hwX509 "huawei.com/mindx/common/x509"
)

// CreateCsr [method] for create csr content
func CreateCsr(keyPath string, commonNamePrefix string, kmcCfg *kmc.SubConfig, san CertSan) ([]byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, priKeyLength)
	if err != nil {
		return nil, errors.New("generate new key for self signed certificate failed: " + err.Error())
	}

	csr, err := getCsr(priv, commonNamePrefix, san)
	if err != nil {
		return nil, err
	}

	if err = saveKeyWithPem(keyPath, priv, kmcCfg); err != nil {
		return nil, errors.New("save self singed key with pem failed: " + err.Error())
	}

	return csr, nil
}

// CreateKubeConfigCsr create csr send to k8s
func CreateKubeConfigCsr(keyPath string, commonName string, kmcCfg *kmc.SubConfig, san CertSan) ([]byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, priKeyLength)
	if err != nil {
		return nil, errors.New("generate new key for self signed certificate failed: " + err.Error())
	}
	csr, err := getKubeConfigCsr(priv, commonName, san)
	if err != nil {
		return nil, err
	}

	if err = saveKeyWithPem(keyPath, priv, kmcCfg); err != nil {
		return nil, errors.New("save self singed key with pem failed: " + err.Error())
	}

	return csr, nil
}

func getKubeConfigCsr(priv *rsa.PrivateKey, commonName string, certSan CertSan) ([]byte, error) {
	template := x509.CertificateRequest{
		Subject: pkix.Name{
			Country:            []string{"kubernetes"},
			Organization:       []string{"kubernetes"},
			OrganizationalUnit: []string{"kubernetes"},
			CommonName:         commonName,
		},
		DNSNames:    certSan.DnsName,
		IPAddresses: certSan.IpAddr,
	}
	csrDer, err := x509.CreateCertificateRequest(rand.Reader, &template, priv)
	if err != nil {
		return nil, errors.New("generate csr for self signed certificate failed: " + err.Error())
	}
	return csrDer, nil
}

func getCsr(priv *rsa.PrivateKey, commonNamePrefix string, certSan CertSan) ([]byte, error) {
	commonNameSuffix, err := envutils.GetUuid()
	if err != nil {
		return nil, errors.New("generate uuid for self signed certificate failed: " + err.Error())
	}

	template := x509.CertificateRequest{
		Subject: pkix.Name{
			Country:            []string{caCountry},
			Organization:       []string{caOrganization},
			OrganizationalUnit: []string{caOrganizationalUnit},
			CommonName:         commonNamePrefix + "-" + commonNameSuffix,
		},
		DNSNames:    certSan.DnsName,
		IPAddresses: certSan.IpAddr,
	}
	csrDer, err := x509.CreateCertificateRequest(rand.Reader, &template, priv)
	if err != nil {
		return nil, errors.New("generate csr for self signed certificate failed: " + err.Error())
	}
	return csrDer, nil
}

// GetCertPair get cert and key pair info
func GetCertPair(certPath, keyPath string, kmcCfg *kmc.SubConfig) (*CaPairInfo, error) {
	pairWithPem, err := GetCertPairForPem(certPath, keyPath, kmcCfg)
	if err != nil {
		return nil, err
	}
	defer utils.ClearSliceByteMemory(pairWithPem.KeyPem)

	caCert, err := hwX509.LoadCertsFromPEM(pairWithPem.CertPem)
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
func GetCertPairForPem(certPath, keyPath string, kmcCfg *kmc.SubConfig) (*CaPairInfoWithPem, error) {
	certBytes, err := fileutils.LoadFile(certPath)
	if certBytes == nil {
		return nil, fmt.Errorf("load cert path [%s] file failed", certPath)
	}
	encryptKeyContent, err := fileutils.LoadFile(keyPath)
	if encryptKeyContent == nil {
		return nil, fmt.Errorf("load key path [%s] file failed: ", keyPath)
	}
	decryptKeyByte, err := kmc.DecryptContent(encryptKeyContent, kmcCfg)
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

// PemWrapCsr code csr to pem type
func PemWrapCsr(der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  pubCsrType,
		Bytes: der,
	})
}

// PemWrapPrivKey code der private key to pem type
func PemWrapPrivKey(priv *rsa.PrivateKey) []byte {
	if priv == nil {
		return nil
	}
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
	defer utils.ClearSliceByteMemory(pm.Bytes)

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
	if err := fileutils.WriteData(certPath, certPem); err != nil {
		return err
	}

	return fileutils.SetPathPermission(certPath, fileutils.Mode400, false, false)
}

func saveKeyWithPem(keyPath string, keyDerBytes *rsa.PrivateKey, kmcCfg *kmc.SubConfig) error {
	keyPem := PemWrapPrivKey(keyDerBytes)
	defer hwX509.PaddingAndCleanSlice(keyPem)
	encryptKeyPem, err := kmc.EncryptContent(keyPem, kmcCfg)
	if err != nil {
		return err
	}
	if err = fileutils.MakeSureDir(keyPath); err != nil {
		return err
	}
	if fileutils.IsExist(keyPath) {
		if err = fileutils.SetPathPermission(keyPath, fileutils.Mode600, false, false); err != nil {
			return err
		}
	}
	defer func() {
		if err := fileutils.SetPathPermission(keyPath, fileutils.Mode400, false, false); err != nil {
			hwlog.RunLog.Error("set path permission 400 error, when save key with pem")
			return
		}
	}()
	hwlog.RunLog.Infof("The key %s has been saved.", keyPath)
	if err := hwX509.OverridePassWdFile(keyPath, encryptKeyPem, fileutils.Mode600); err != nil {
		return err
	}

	return fileutils.SetPathPermission(keyPath, fileutils.Mode400, false, false)
}

// GetCertContent return a cert file's content by path after checking whether the cert path can be loaded and parsed.
func GetCertContent(path string) ([]byte, error) {
	certBytes, err := fileutils.LoadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load cert file from path [%s] failed, %v", path, err)
	}
	_, err = hwX509.NewCaChainMgr(certBytes)
	if err != nil {
		return nil, fmt.Errorf("parse cert failed, %v", err)
	}
	return certBytes, nil
}

// GetCertContentWithBackup return a cert file's content by path after checking.
// And when failed at first, it will try restore file from backup and try again.
// when success at first, it will try update backup file by current path.
func GetCertContentWithBackup(path string) ([]byte, error) {
	cert, getErr := GetCertContent(path)
	if getErr == nil {
		if backErr := backuputils.BackUpFiles(path); backErr != nil {
			hwlog.RunLog.Warnf("back up cert [%s] failed, %v", path, backErr)
		}
		return cert, nil
	}
	hwlog.RunLog.Warnf("load root cert [%s] failed, %v, try restore from backup", path, getErr)
	if restoreErr := backuputils.RestoreFiles(path); restoreErr != nil {
		hwlog.RunLog.Errorf("try restore [%s] from backup failed, %v", path, restoreErr)
		return nil, getErr
	}
	return GetCertContent(path)
}

// GetKeyContent return a key file's content by path after checking and decrypting.
func GetKeyContent(path string, kmcCfg *kmc.SubConfig) ([]byte, error) {
	keyBytes, err := fileutils.LoadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load key file from path [%s] failed", path)
	}
	decryptKeyByte, err := kmc.DecryptContent(keyBytes, kmcCfg)
	if err != nil {
		return nil, fmt.Errorf("load key file from path [%s] failed, decrypt key content failed", path)
	}
	return decryptKeyByte, nil
}

// GetKeyContentWithBackup return a key file's content by path after checking and decrypting.
func GetKeyContentWithBackup(path string, kmcCfg *kmc.SubConfig) ([]byte, error) {
	key, getErr := GetKeyContent(path, kmcCfg)
	if getErr == nil {
		if backErr := backuputils.BackUpFiles(path); backErr != nil {
			hwlog.RunLog.Warnf("back up key [%s] failed", path)
		}
		return key, nil

	}
	hwlog.RunLog.Warnf("load root key [%s] failed, %v, try restore from backup", path, getErr)
	if restoreErr := backuputils.RestoreFiles(path); restoreErr != nil {
		hwlog.RunLog.Errorf("try restore [%s] from backup failed", path)
		return nil, getErr
	}
	return GetKeyContent(path, kmcCfg)
}

// GetCertPairForPemWithBackup [method] for get cert pair with backup
func GetCertPairForPemWithBackup(certPath, keyPath string, kmcCfg *kmc.SubConfig) (*CaPairInfoWithPem, error) {
	certBytes, err := GetCertContentWithBackup(certPath)
	if certBytes == nil {
		return nil, fmt.Errorf("load cert path [%s] file failed", certPath)
	}
	decryptKeyByte, err := GetKeyContentWithBackup(keyPath, kmcCfg)
	if err != nil {
		return nil, fmt.Errorf("load and decrypt key content failed: %v", err)
	}
	return &CaPairInfoWithPem{CertPem: certBytes, KeyPem: decryptKeyByte}, nil
}

// GetCrlContentWithBackup parses crl in PEM format and returns crl content. This function will try to restore crl from
// backup file if main crl file is broken.
func GetCrlContentWithBackup(path string) ([]byte, error) {
	crl, getErr := GetCrlContent(path)
	if getErr == nil {
		if backErr := backuputils.BackUpFiles(path); backErr != nil {
			hwlog.RunLog.Warnf("back up crl [%s] failed, %v", path, backErr)
		}
		return crl, nil
	}
	hwlog.RunLog.Warnf("load root crl [%s] failed, %v, try restore from backup", path, getErr)
	if restoreErr := backuputils.RestoreFiles(path); restoreErr != nil {
		hwlog.RunLog.Errorf("try restore [%s] from backup failed, %v", path, restoreErr)
		return nil, getErr
	}
	return GetCrlContent(path)
}

// GetCrlContent parses crl in PEM format and returns crl content
func GetCrlContent(path string) ([]byte, error) {
	crlContent, err := fileutils.LoadFile(path)
	if err != nil {
		return nil, err
	}

	if _, err := hwX509.NewCrlMgr(crlContent); err != nil {
		return nil, fmt.Errorf("new crl Chain Mgr failed: %s", err.Error())
	}
	return crlContent, nil
}
