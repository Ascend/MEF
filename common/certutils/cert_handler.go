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
	"huawei.com/mindx/common/utils"
	hwCertMgr "huawei.com/mindx/common/x509"
	"huawei.com/mindxedge/base/common"
)

const (
	// priKeyLength private key length
	priKeyLength = 4096
	// validationYearCA root ca validate year
	validationYearCA = 10
	// validationYearCert service Cert validate year
	validationYearCert = 10
	// validationMonth Cert validate month
	validationMonth = 0
	// validationDay Cert validate day
	validationDay = 0
	// bigIntSize server_number
	bigIntSize = 2022
	// caCountry issue country
	caCountry = "CN"
	// caOrganization issue organization
	caOrganization = "Huawei"
	// caOrganizationalUnit issue unit
	caOrganizationalUnit = "Ascend"
	// caCommonName issue name
	caCommonName = "MEF"
	// pubCertType Cert type
	pubCertType = "CERTIFICATE"
	// privKeyType Cert key type
	privKeyType = "RSA PRIVATE KEY"
	// fileMode Cert file mode
	fileMode = 0600
)

// RootCertMgr define root cert manager struct
type RootCertMgr struct {
	rootCaPath  string
	rootKeyPath string
	commonName  string
	kmcCfg      *common.KmcCfg
}

// CaPairInfo define cert and key pair info struct
type CaPairInfo struct {
	Cert   *x509.Certificate
	PriKey *rsa.PrivateKey
}

// GetCertPair get cert and key pair info
func GetCertPair(certPath, keyPath string, kmcCfg *common.KmcCfg) (*CaPairInfo, error) {
	rootCa, err := utils.LoadFile(certPath)
	if err != nil {
		return nil, errors.New("load cert file failed: " + err.Error())
	}
	caCert, err := hwCertMgr.LoadCertsFromPEM(rootCa)
	if err != nil {
		return nil, errors.New("decode cert form pem failed: " + err.Error())
	}
	encryptKeyContent, err := utils.LoadFile(keyPath)
	if err != nil {
		return nil, errors.New("load key file failed: " + err.Error())
	}
	decryptKeyByte, err := common.DecryptContent(encryptKeyContent, kmcCfg)
	if err != nil {
		return nil, errors.New("decrypt key content failed: " + err.Error())
	}
	defer hwCertMgr.PaddingAndCleanSlice(decryptKeyByte)
	caPrivate := PemUnwrapPrivKey(decryptKeyByte)
	if caPrivate == nil {
		return nil, errors.New("unwrap a private key pem failed")
	}
	var caPairInfo = &CaPairInfo{
		Cert:   caCert,
		PriKey: caPrivate,
	}
	return caPairInfo, nil
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

	cer := rcm.makeServiceCertificate(srvCsr.Subject)
	certBytes, err := x509.CreateCertificate(rand.Reader, cer, rootCaPair.Cert, srvCsr.PublicKey, rootCaPair.PriKey)
	if err != nil {
		return nil, errors.New("create service certificate failed: " + err.Error())
	}
	return certBytes, nil
}

func (rcm *RootCertMgr) makeServiceCertificate(subject pkix.Name) *x509.Certificate {
	now := time.Now()
	return &x509.Certificate{
		SerialNumber: big.NewInt(bigIntSize),
		Subject:      subject,
		NotBefore:    now,
		NotAfter:     now.AddDate(validationYearCert, validationMonth, validationDay),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{subject.CommonName},
	}
}

func (rcm *RootCertMgr) getRootCaCsr() *x509.Certificate {
	now := time.Now()
	csr := &x509.Certificate{
		SerialNumber: big.NewInt(bigIntSize),
		Subject: pkix.Name{
			Country:            []string{caCountry},
			Organization:       []string{caOrganization},
			OrganizationalUnit: []string{caOrganizationalUnit},
			CommonName:         caCommonName + rcm.commonName,
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

// SelfSignCert self singed cert struct
type SelfSignCert struct {
	rootCertMgr *RootCertMgr
	svcCertPath string
	svcKeyPath  string
	commonName  string
}

// CreateSignCert create a new singed cert for root ca and service cert
func (sc *SelfSignCert) CreateSignCert() error {
	if _, err := sc.rootCertMgr.GetRootCaPair(); err != nil {
		if _, err := sc.rootCertMgr.NewRootCa(); err != nil {
			return errors.New("new root ca pair failed: " + err.Error())
		}
	}
	priv, err := rsa.GenerateKey(rand.Reader, priKeyLength)
	if err != nil {
		return errors.New("generate new key for self signed certificate failed: " + err.Error())
	}

	csr, err := sc.getCsr(priv)
	if err != nil {
		return err
	}

	certBytes, err := sc.rootCertMgr.IssueServiceCert(csr)
	if err != nil {
		return err
	}

	err = saveCertWithPem(sc.svcCertPath, certBytes)
	if err != nil {
		return errors.New("save self singed cert with pem failed: " + err.Error())
	}

	err = saveKeyWithPem(sc.svcKeyPath, priv, sc.rootCertMgr.kmcCfg)
	if err != nil {
		return errors.New("save self singed key with pem failed: " + err.Error())
	}
	return nil
}

func (sc *SelfSignCert) getCsr(priv *rsa.PrivateKey) ([]byte, error) {
	template := x509.CertificateRequest{Subject: pkix.Name{CommonName: sc.commonName}}
	csrDer, err := x509.CreateCertificateRequest(rand.Reader, &template, priv)
	if err != nil {
		return nil, errors.New("generate csr for self signed certificate failed: " + err.Error())
	}
	return csrDer, nil
}
