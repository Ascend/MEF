// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package certmanager issues service cert for edge-installer
package certmanager

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"huawei.com/mindx/common/hwlog"
	"math/big"
	"time"
)

func (rootCAs *RootCAs) issueService(csrByte []byte) []byte {
	certs := rootCAs.RootCaValidEdge

	csr, err := x509.ParseCertificateRequest(csrByte)
	if err != nil {
		hwlog.RunLog.Errorf("parse certificate request failed, error: %v", err)
		return nil
	}

	template := geneService(csr.Subject)
	serCerBytes, err := x509.CreateCertificate(rand.Reader, template, certs.RootCA, csr.PublicKey, certs.CaPriKey)
	if err != nil {
		hwlog.RunLog.Errorf("create certificate failed, error: %v", err)
		return nil
	}

	return serCerBytes
}

func geneService(subject pkix.Name) *x509.Certificate {
	now := time.Now()
	return &x509.Certificate{
		SerialNumber: big.NewInt(NewInt),
		Subject:      subject,
		NotBefore:    now,
		NotAfter:     now.AddDate(ValidationYear, ValidationMonth, ValidationDay),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{subject.CommonName},
	}
}
