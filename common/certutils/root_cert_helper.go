// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package certutils Cert mgr
package certutils

import (
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"math/big"
	"time"

	"huawei.com/mindx/common/rand"

	"huawei.com/mindxedge/base/common"
)

// RootCertMgr define root cert manager struct
type RootCertMgr struct {
	rootCaPath  string
	rootKeyPath string
	commonName  string
	kmcCfg      *common.KmcCfg
}

// InitRootCertMgr init root cert manager
func InitRootCertMgr(caPath, keyPath, commonName string, kmcCfg *common.KmcCfg) *RootCertMgr {
	if kmcCfg == nil {
		kmcCfg = common.GetDefKmcCfg()
	}
	var mgr = &RootCertMgr{
		rootCaPath:  caPath,
		rootKeyPath: keyPath,
		commonName:  commonName,
		kmcCfg:      kmcCfg,
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
	rootCsr := rcm.getRootCaCsr()
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
	err = saveCertWithPem(rcm.rootCaPath, rootCaBytes)
	if err != nil {
		return nil, errors.New("save root cert with pem failed: " + err.Error())
	}

	err = saveKeyWithPem(rcm.rootKeyPath, caPriKey, rcm.kmcCfg)
	if err != nil {
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

	srvCsr, err := x509.ParseCertificateRequest(csr)
	if err != nil {
		return nil, errors.New("parse certificate request for csr failed: " + err.Error())
	}

	cer := rcm.makeServiceCertificate(srvCsr)
	certBytes, err := x509.CreateCertificate(rand.Reader, cer, rootCaPair.Cert, srvCsr.PublicKey, rootCaPair.PriKey)
	if err != nil {
		return nil, errors.New("create service certificate failed: " + err.Error())
	}
	return certBytes, nil
}

func (rcm *RootCertMgr) makeServiceCertificate(csr *x509.CertificateRequest) *x509.Certificate {
	now := time.Now().UTC()
	if oneDayAgo, err := time.ParseDuration(OneDayAgo); err == nil {
		now = now.Add(oneDayAgo)
	}
	return &x509.Certificate{
		SerialNumber: big.NewInt(bigIntSize),
		Subject:      csr.Subject,
		NotBefore:    now,
		NotAfter:     now.AddDate(validationYearCert, validationMonth, validationDay),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		DNSNames:     csr.DNSNames,
		IPAddresses:  csr.IPAddresses,
	}
}

func (rcm *RootCertMgr) getRootCaCsr() *x509.Certificate {
	now := time.Now().UTC()
	if oneDayAgo, err := time.ParseDuration(OneDayAgo); err == nil {
		now = now.Add(oneDayAgo)
	}
	csr := &x509.Certificate{
		SerialNumber: big.NewInt(bigIntSize),
		Subject: pkix.Name{
			Country:            []string{caCountry},
			Organization:       []string{caOrganization},
			OrganizationalUnit: []string{caOrganizationalUnit},
			CommonName:         caCommonName + "-" + rcm.commonName,
		},
		NotBefore:             now,
		NotAfter:              now.AddDate(validationYearCA, validationMonth, validationDay),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	return csr
}
