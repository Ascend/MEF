// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package certmanager generates root ca and save
package certmanager

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"huawei.com/mindxedge/base/common"

	"encoding/pem"
	"errors"
	"huawei.com/mindxedge/base/modulemanager/model"
	"math/big"
	"os"
	"path"
	"time"

	"huawei.com/mindx/common/hwlog"
	pubX509 "huawei.com/mindx/common/x509"
)

// RootCAs root cas info
type RootCAs struct {
	RootCaValidCenter *rootCaInfo
	RootCaValidEdge   *rootCaInfo
	certDir           string
}

// rootCaInfo includes private key and bytes of cert
type rootCaInfo struct {
	RootCA           *x509.Certificate
	CaPriKey         *rsa.PrivateKey
	CaBytes          []byte
	rootCaName       string
	rootCaBackUpName string
}

// NewRootCAs create RootCAs instance
func NewRootCAs() *RootCAs {
	certDir, ok := getCertDir()
	if !ok {
		hwlog.RunLog.Error("get cert directory failed")
		return nil
	}
	rootCaValidCenter := newRootCaValidCenter()
	rootCaValidEdge := newRootCaValidEdge()
	return &RootCAs{
		certDir:           certDir,
		RootCaValidCenter: rootCaValidCenter,
		RootCaValidEdge:   rootCaValidEdge,
	}
}

func newRootCaValidCenter() *rootCaInfo {
	return &rootCaInfo{
		rootCaName:       RootCaNameValidCenter,
		rootCaBackUpName: RootCaBackUpNameValidCenter,
	}
}

func newRootCaValidEdge() *rootCaInfo {
	return &rootCaInfo{
		rootCaName:       RootCaNameValidEdge,
		rootCaBackUpName: RootCaBackUpNameValidEdge,
	}
}

func getCertDir() (string, bool) {
	currentPath, ok := common.GetEdgeMgrWorkPath()
	if !ok {
		hwlog.RunLog.Error("get the edge-manager work path failed")
		return "", false
	}

	certDir := path.Join(currentPath, CertMgrPathName)
	return certDir, true
}

// GenerateRootCA generates root cas for validating center and edge
func (rootCAs *RootCAs) GenerateRootCA() {
	if err := rootCAs.genRootCaAndSave(ValidCenter); err != nil {
		hwlog.RunLog.Errorf("generate root ca validating center failed, error: %v", err)
		return
	}
	if err := rootCAs.genRootCaAndSave(ValidEdge); err != nil {
		hwlog.RunLog.Errorf("generate root ca validating edge failed, error: %v", err)
		return
	}
	return
}

func (rootCAs *RootCAs) genRootCaAndSave(caInfo string) error {
	certs := newCertificates()
	if certs == nil {
		return errors.New("new certificates failed")
	}

	var rootCaName, rootCaBackUpName string
	switch caInfo {
	case ValidCenter:
		rootCaName = rootCAs.RootCaValidCenter.rootCaName
		rootCaBackUpName = rootCAs.RootCaValidCenter.rootCaBackUpName
	case ValidEdge:
		rootCaName = rootCAs.RootCaValidEdge.rootCaName
		rootCaBackUpName = rootCAs.RootCaValidEdge.rootCaBackUpName
	default:
		return errors.New("caInfo invalid")
	}
	if err := rootCAs.writeRootCaToFile(rootCaName, rootCaBackUpName, certs); err != nil {
		return err
	}
	return nil
}

func newCertificates() *rootCaInfo {
	capriKey, caBytes, err := prepareRootCAInfo()
	if capriKey == nil || err != nil {
		hwlog.RunLog.Error("new root ca failed")
		return nil
	}

	ca, err := x509.ParseCertificate(caBytes)
	if err != nil {
		hwlog.RunLog.Errorf("failed to parse root CA certificate: %v", err)
		return nil
	}

	return &rootCaInfo{
		RootCA:   ca,
		CaPriKey: capriKey,
		CaBytes:  caBytes,
	}
}

func prepareRootCAInfo() (*rsa.PrivateKey, []byte, error) {
	priKey, err := rsa.GenerateKey(rand.Reader, PrivateKeyBits)
	if err != nil {
		hwlog.RunLog.Errorf("generate private key failed, error: %v", err)
		return nil, nil, err
	}

	now := time.Now()
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(NewInt),
		Subject: pkix.Name{
			CommonName:         common.MEF,
			Country:            []string{CaCountry},
			Organization:       []string{CaOrganization},
			OrganizationalUnit: []string{CaOrganizationalUnit},
		},
		NotBefore:             now,
		NotAfter:              now.AddDate(ValidationYear, ValidationMonth, ValidationDay),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &priKey.PublicKey, priKey)
	if err != nil {
		hwlog.RunLog.Errorf("create certificate failed, error: %v", err)
		return nil, nil, err
	}
	return priKey, caBytes, nil
}

func (rootCAs *RootCAs) writeRootCaToFile(rootCaName, rootCaBackUpName string, certs *rootCaInfo) error {
	if certs == nil {
		return errors.New("certs is nil")
	}

	var pemByte []byte
	var err error

	rootCaFileName := path.Join(rootCAs.certDir, rootCaName)
	if pemByte, err = generateCaFile(rootCaFileName, CERTIFICATE, certs); err != nil {
		hwlog.RunLog.Error("generate ca file failed")
		return err
	}

	rootCaBackUpFileName := path.Join(rootCAs.certDir, rootCaBackUpName)
	buInstance, err := pubX509.NewBKPInstance(pemByte, rootCaFileName, rootCaBackUpFileName)
	if err != nil {
		hwlog.RunLog.Errorf("back up root ca failed, error: %v", err)
		return err
	}
	if err = buInstance.WriteToDisk(CertFileMode, false); err != nil {
		return err
	}

	return nil
}

func generateCaFile(path, pemType string, cert *rootCaInfo) ([]byte, error) {
	pemBuf := new(bytes.Buffer)
	if err := pem.Encode(pemBuf, &pem.Block{
		Type:  pemType,
		Bytes: cert.CaBytes,
	}); err != nil {
		hwlog.RunLog.Errorf("PEM encoding error: %v", err)
		return nil, err
	}
	pemByte := pemBuf.Bytes()
	if err := os.WriteFile(path, pemByte, CertFileMode); err != nil {
		hwlog.RunLog.Errorf("write ca file error: %v", err)
		return nil, err
	}

	return pemByte, nil
}

type exportRootCAReq struct {
	Path string `json:"path"`
}

func (rootCAs *RootCAs) exportRootCaValidCenter(msg *model.Message) { // 后续需要将二进制流返给restful进行下载。不指定下载路径
	rootCAExport, ok := msg.GetContent().(exportRootCAReq)
	if !ok {
		hwlog.RunLog.Error("convert to rootCAExport failed")
		return
	}
	caFile := path.Join(rootCAExport.Path, "rootCA.crt")
	cert := rootCAs.RootCaValidCenter
	err := os.WriteFile(caFile, cert.CaBytes, CertFileMode)
	if err != nil {
		hwlog.RunLog.Errorf("write to file error: %v", err)
		return
	}
	return
}
