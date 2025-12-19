// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package certutils Cert mgr
package certutils

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"math/big"
	"time"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/rand"
)

// RootCertMgr define root cert manager struct
type RootCertMgr struct {
	kmcCfg           *kmc.SubConfig
	rootCaPath       string
	rootKeyPath      string
	commonNamePrefix string
}

// InitRootCertMgr init root cert manager
func InitRootCertMgr(caPath, keyPath, commonNamePrefix string, kmcCfg *kmc.SubConfig) *RootCertMgr {
	if kmcCfg == nil {
		kmcCfg = kmc.GetDefKmcCfg()
	}
	var mgr = &RootCertMgr{
		rootCaPath:       caPath,
		rootKeyPath:      keyPath,
		commonNamePrefix: commonNamePrefix,
		kmcCfg:           kmcCfg,
	}
	return mgr
}

// GetRootCaPair get root ca and key pair
func (rcm *RootCertMgr) GetRootCaPair() (*CaPairInfo, error) {
	return GetCertPair(rcm.rootCaPath, rcm.rootKeyPath, rcm.kmcCfg)
}

// NewRootCa new root ca
func (rcm *RootCertMgr) NewRootCa() (*CaPairInfo, error) {
	caPriKey, err := rsa.GenerateKey(rand.Reader, priKeyLength)
	if err != nil {
		return nil, errors.New("generate key failed: " + err.Error())
	}
	rootCsr, err := rcm.getRootCaCsr()
	if err != nil {
		return nil, errors.New("generate csr failed: " + err.Error())
	}

	pubKeyBytes := x509.MarshalPKCS1PublicKey(&caPriKey.PublicKey)
	pubKeySha256 := sha256.Sum256(pubKeyBytes)
	rootCsr.SubjectKeyId = pubKeySha256[:]
	rootCsr.AuthorityKeyId = pubKeySha256[:]

	rootCaBytes, err := x509.CreateCertificate(rand.Reader, rootCsr, rootCsr, &caPriKey.PublicKey, caPriKey)
	if err != nil {
		return nil, errors.New("CreateCertificate root ca failed: " + err.Error())
	}
	rootCaCert, err := x509.ParseCertificate(rootCaBytes)
	if err != nil {
		return nil, errors.New("ParseCertificate root ca failed: " + err.Error())
	}
	var rootCaInfo = &CaPairInfo{
		Cert:   rootCaCert,
		PriKey: caPriKey,
	}
	if err = saveCertWithPem(rcm.rootCaPath, rootCaBytes); err != nil {
		return nil, errors.New("save root cert with pem failed: " + err.Error())
	}

	if err = saveKeyWithPem(rcm.rootKeyPath, caPriKey, rcm.kmcCfg); err != nil {
		return nil, errors.New("save root key with pem failed: " + err.Error())
	}

	return rootCaInfo, nil
}

// IssueServiceCert issues a service certificate with csr
func (rcm *RootCertMgr) IssueServiceCert(csr []byte) ([]byte, error) {
	rootCaPair, err := rcm.GetRootCaPair()
	if err != nil {
		return nil, errors.New("get root ca pair failed: " + err.Error())
	}
	return rcm.issueServiceCertByCaPair(rootCaPair, csr)
}

func (rcm *RootCertMgr) issueServiceCertByCaPair(rootCaPair *CaPairInfo, csr []byte) ([]byte, error) {
	if rootCaPair == nil {
		return nil, errors.New("root ca pair is nil")
	}
	srvCsr, err := x509.ParseCertificateRequest(csr)
	if err != nil {
		return nil, errors.New("parse certificate request for csr failed: " + err.Error())
	}

	if err := srvCsr.CheckSignature(); err != nil {
		return nil, errors.New("check signature for csr failed: " + err.Error())
	}

	cer, err := rcm.makeServiceCertificate(srvCsr)
	if err != nil {
		return nil, errors.New("make service certificate failed: " + err.Error())
	}

	servPubKeyBytes, err := x509.MarshalPKIXPublicKey(srvCsr.PublicKey)
	if err != nil {
		return nil, errors.New("parse srv pub key failed: " + err.Error())
	}
	pubKeySha256 := sha256.Sum256(servPubKeyBytes)
	cer.SubjectKeyId = pubKeySha256[:]

	rootPubKeyBytes, err := x509.MarshalPKIXPublicKey(rootCaPair.Cert.PublicKey)
	if err != nil {
		return nil, errors.New("parse root certificate pub key failed: " + err.Error())
	}
	pubKeySha256 = sha256.Sum256(rootPubKeyBytes)
	cer.AuthorityKeyId = pubKeySha256[:]

	certBytes, err := x509.CreateCertificate(rand.Reader, cer, rootCaPair.Cert, srvCsr.PublicKey, rootCaPair.PriKey)
	if err != nil {
		return nil, errors.New("create service certificate failed: " + err.Error())
	}
	return certBytes, nil
}

func (rcm *RootCertMgr) makeServiceCertificate(csr *x509.CertificateRequest) (*x509.Certificate, error) {
	now := time.Now().UTC()
	if oneDayAgo, err := time.ParseDuration(OneDayAgo); err == nil {
		now = now.Add(oneDayAgo)
	}
	sn, err := rcm.createRandomSn()
	if err != nil {
		return nil, err
	}

	return &x509.Certificate{
		SerialNumber: sn,
		Subject:      csr.Subject,
		NotBefore:    now,
		NotAfter:     now.AddDate(validationYearCert, validationMonth, validationDay),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		DNSNames:     csr.DNSNames,
		IPAddresses:  csr.IPAddresses,
	}, nil
}

func (rcm *RootCertMgr) getRootCaCsr() (*x509.Certificate, error) {
	now := time.Now().UTC()
	if oneDayAgo, err := time.ParseDuration(OneDayAgo); err == nil {
		now = now.Add(oneDayAgo)
	}
	sn, err := rcm.createRandomSn()
	if err != nil {
		return nil, err
	}
	commonNameSuffix, err := envutils.GetUuid()
	if err != nil {
		return nil, errors.New("generate uuid for self signed certificate failed: " + err.Error())
	}

	csr := &x509.Certificate{
		SerialNumber: sn,
		Subject: pkix.Name{
			Country:            []string{caCountry},
			Organization:       []string{caOrganization},
			OrganizationalUnit: []string{caOrganizationalUnit},
			CommonName:         rcm.commonNamePrefix + "-" + commonNameSuffix,
		},
		NotBefore:             now,
		NotAfter:              now.AddDate(validationYearCA, validationMonth, validationDay),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	return csr, nil
}

func (rcm *RootCertMgr) createRandomSn() (*big.Int, error) {
	const snSize = 20
	random := make([]byte, snSize)
	_, err := rand.Read(random)
	if err != nil {
		return nil, fmt.Errorf("read random nums failed: %s", err.Error())
	}

	return new(big.Int).SetBytes(random), nil
}

// GetRootCaPairWithBackup get root ca and key pair, will create backup files in same path,
// and will try again if failed at first time after restoring files from backup.
func (rcm *RootCertMgr) GetRootCaPairWithBackup() (*CaPairInfo, error) {
	pair, getErr := GetCertPair(rcm.rootCaPath, rcm.rootKeyPath, rcm.kmcCfg)
	if getErr == nil {
		if backErr := rcm.backupCaAndKey(); backErr != nil {
			hwlog.RunLog.Warnf("back up ca or key failed, %v", backErr)
		}
		return pair, nil
	}

	hwlog.RunLog.Warnf("check cert failed, %v, try restore it from backup", getErr)
	if restoreErr := rcm.restoreCaAndKey(); restoreErr != nil {
		return nil, fmt.Errorf("restore ca or key from backup failed, %v", restoreErr)
	}

	pair, getErr = GetCertPair(rcm.rootCaPath, rcm.rootKeyPath, rcm.kmcCfg)
	if getErr != nil {
		return nil, fmt.Errorf("get root ca pair failed after recovery from backup, %v", getErr)
	}
	return pair, nil
}

// NewRootCaWithBackup new root ca, then generate a backup file in same path which end with .backup suffix
func (rcm *RootCertMgr) NewRootCaWithBackup() (*CaPairInfo, error) {
	rootCaInfo, err := rcm.NewRootCa()
	if err != nil {
		return nil, err
	}
	if backErr := rcm.backupCaAndKey(); backErr != nil {
		hwlog.RunLog.Warnf("back up ca or key failed, %v", backErr)
	}
	return rootCaInfo, nil
}

func (rcm *RootCertMgr) backupCaAndKey() error {
	return backuputils.BackUpFiles(rcm.rootCaPath, rcm.rootKeyPath)
}

func (rcm *RootCertMgr) restoreCaAndKey() error {
	if _, checkErr := GetCertContent(rcm.rootCaPath); checkErr != nil {
		if restoreErr := backuputils.RestoreFiles(rcm.rootCaPath); restoreErr != nil {
			return restoreErr
		}
	}
	if checkErr := CheckKey(rcm.rootKeyPath, rcm.kmcCfg); checkErr != nil {
		if restoreErr := backuputils.RestoreFiles(rcm.rootKeyPath); restoreErr != nil {
			return restoreErr
		}
	}
	return nil
}

// IssueServiceCertWithBackup issues a service certificate with csr
func (rcm *RootCertMgr) IssueServiceCertWithBackup(csr []byte) ([]byte, error) {
	rootCaPair, err := rcm.GetRootCaPairWithBackup()
	if err != nil {
		return nil, fmt.Errorf("get root ca pair failed:  %v", err)
	}
	return rcm.issueServiceCertByCaPair(rootCaPair, csr)
}
