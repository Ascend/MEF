//  Copyright(c) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

package x509

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
)

const (
	certBegin    = "-----BEGIN CERTIFICATE-----"
	maxCertCount = 10
)

// CaChainMgr is the struct to manager a ca chain
type CaChainMgr struct {
	certs []*x509.Certificate
}

// CaChainCheckOptions is the struct to define ca chain check options. Using default value is strongly suggested.
type CaChainCheckOptions struct {
	// AllowMiddleStrengthRsaPublicKey a 2048-bit rsa public key is considered as middle strength.
	// Turn on this option only when you import third-party certificates.
	AllowMiddleStrengthRsaPublicKey bool
}

// NewCaChainMgr is used to create and init a CaChainMgr
// the content support PEM format certs
func NewCaChainMgr(content []byte) (*CaChainMgr, error) {
	caMgr := CaChainMgr{}

	err := caMgr.parseCerts(content)
	if err != nil {
		return nil, err
	}

	return &caMgr, nil
}

// NewCaChainMgrFromDer is used to create and init a CaChainMgr
// the content support DER format certs
func NewCaChainMgrFromDer(content []byte) (*CaChainMgr, error) {
	caMgr := CaChainMgr{}

	err := caMgr.parseDerCerts(content)
	if err != nil {
		return nil, err
	}

	return &caMgr, nil
}

// GetCerts get certs
func (ccm *CaChainMgr) GetCerts() []*x509.Certificate {
	return ccm.certs
}

func (ccm *CaChainMgr) parseCerts(content []byte) error {
	if len(content) <= 0 {
		return errors.New("content is empty")
	}

	var block *pem.Block
	var derCerts []byte
	splittedCerts := strings.Split(string(content), certBegin)
	if len(splittedCerts) <= 1 {
		return errors.New("no PEM cert contains")
	}

	for _, singleCert := range splittedCerts[1:] {
		sigContent := []byte(certBegin + singleCert)
		block, _ = pem.Decode(sigContent)
		if block == nil {
			return errors.New("parse cert failed")
		}
		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			return errors.New("invalid cert bytes")
		}
		derCerts = append(derCerts, block.Bytes...)
	}

	certs, err := x509.ParseCertificates(derCerts)
	if err != nil {
		return fmt.Errorf("parse Certificates failed: %s", err.Error())
	}
	if len(certs) > maxCertCount {
		return fmt.Errorf("certs count %d exceed limitation %d", len(certs), maxCertCount)
	}

	ccm.certs = certs
	return nil
}

func (ccm *CaChainMgr) parseDerCerts(content []byte) error {
	if len(content) <= 0 {
		return errors.New("content is empty")
	}

	certs, err := x509.ParseCertificates(content)
	if err != nil {
		return fmt.Errorf("parse Certificates failed: %s", err.Error())
	}
	if len(certs) <= 0 {
		return fmt.Errorf("unsupported cert type, no cert is parsed out")
	}
	if len(certs) > maxCertCount {
		return fmt.Errorf("certs count %d exceed limitation %d", len(certs), maxCertCount)
	}

	ccm.certs = certs
	return nil
}

// CheckCertChain is the main func to check entire cert chain
func (ccm *CaChainMgr) CheckCertChain(chainOptions ...CaChainCheckOptions) error {
	var chainOpts CaChainCheckOptions
	if len(chainOptions) > 0 {
		chainOpts = chainOptions[0]
	}
	opts, err := ccm.CreateCertsPool(false)
	if err != nil {
		return err
	}

	if err = ccm.checkCertsWithPool(opts, chainOpts); err != nil {
		return err
	}

	return nil
}

// CheckCertPool is the func to check the entire cert pool
func (ccm *CaChainMgr) CheckCertPool(chainOptions ...CaChainCheckOptions) error {
	var chainOpts CaChainCheckOptions
	if len(chainOptions) > 0 {
		chainOpts = chainOptions[0]
	}
	opts, err := ccm.CreateCertsPool(true)
	if err != nil {
		return err
	}

	if err = ccm.checkCertsWithPool(opts, chainOpts); err != nil {
		return err
	}

	return nil
}

// CreateCertsPool is used to create a cert pool by the certs it manages
func (ccm *CaChainMgr) CreateCertsPool(allowMultipleRootCAs bool) (x509.VerifyOptions, error) {
	opts := x509.VerifyOptions{
		Roots:         x509.NewCertPool(),
		Intermediates: x509.NewCertPool(),
	}
	var containsCa bool

	for _, cert := range ccm.certs {
		if ccm.ifRootCaCerts(*cert) {
			if containsCa && !allowMultipleRootCAs {
				return opts, errors.New("the cert chain contains more than 1 root ca")
			}
			opts.Roots.AddCert(cert)
			containsCa = true
		} else {
			opts.Intermediates.AddCert(cert)
		}
	}

	if !containsCa {
		return opts, errors.New("the cert chain does not contain any root ca")
	}

	return opts, nil
}

func (ccm *CaChainMgr) checkCertsWithPool(opts x509.VerifyOptions, chainOpts CaChainCheckOptions) error {
	for _, cert := range ccm.certs {
		if _, err := cert.Verify(opts); err != nil {
			hwlog.RunLog.Info("now is checking cert:")
			hwlog.RunLog.Infof("    ---issuer: %s", cert.Issuer.String())
			hwlog.RunLog.Infof("    ---subject: %s", cert.Subject.String())
			return fmt.Errorf("cert chain contains unauthorized certs: %s", err.Error())
		}

		if err := ccm.checkSingleCerts(cert, chainOpts); err != nil {
			return fmt.Errorf("check cert failed: %s", err.Error())
		}
	}

	return nil
}

func (ccm *CaChainMgr) checkSingleCerts(cert *x509.Certificate, chainOpts CaChainCheckOptions) error {
	if err := CheckCaExtension(cert); err != nil {
		return err
	}

	if err := CheckValidityPeriod(cert, false); err != nil {
		return err
	}

	if err := CheckSignatureAlgorithm(cert); err != nil {
		return err
	}

	if err := CheckPubKeyLength(cert, chainOpts.AllowMiddleStrengthRsaPublicKey); err != nil {
		return err
	}

	return nil
}

func (ccm *CaChainMgr) ifRootCaCerts(cert x509.Certificate) bool {
	if err := cert.CheckSignature(cert.SignatureAlgorithm, cert.RawTBSCertificate, cert.Signature); err != nil {
		return false
	}

	return true
}

// CheckCertsOverdue check if certs is within the allowed expiration time
func (ccm *CaChainMgr) CheckCertsOverdue(overdueTime int) error {
	for _, cert := range ccm.certs {
		if err := CheckValidityPeriodWithError(cert, overdueTime); err != nil {
			return err
		}
	}

	return nil
}

// CheckCertsChainReturnContent is used to check a cert chain file and returns its content
func CheckCertsChainReturnContent(path string, chainOpts ...CaChainCheckOptions) ([]byte, error) {
	caContent, err := fileutils.LoadFile(path)
	if err != nil {
		return nil, err
	}

	chainMgr, err := NewCaChainMgr(caContent)
	if err != nil {
		return nil, fmt.Errorf("new ca Chain Mgr failed: %s", err.Error())
	}

	if err = chainMgr.CheckCertChain(chainOpts...); err != nil {
		return nil, fmt.Errorf("check cert chain failed: %s", err.Error())
	}

	return caContent, nil
}

// GetCerts is the func to get the all cert structures from a cert path
func GetCerts(path string) ([]*x509.Certificate, error) {
	caContent, err := fileutils.LoadFile(path)
	if err != nil {
		return nil, err
	}

	chainMgr, err := NewCaChainMgr(caContent)
	if err != nil {
		return nil, fmt.Errorf("new ca Chain Mgr failed: %s", err.Error())
	}

	return chainMgr.GetCerts(), nil
}

// CheckDerCertChain is used to check a cert chain file from DER cert content
func CheckDerCertChain(content []byte, chainOpts ...CaChainCheckOptions) error {
	chainMgr, err := NewCaChainMgrFromDer(content)
	if err != nil {
		return fmt.Errorf("new ca Chain Mgr failed: %s", err.Error())
	}

	if err = chainMgr.CheckCertChain(chainOpts...); err != nil {
		return fmt.Errorf("check cert chain failed: %s", err.Error())
	}

	return nil
}

// CheckPemCertChain is used to check a cert chain file from PEM cert content
func CheckPemCertChain(content []byte, chainOpts ...CaChainCheckOptions) error {
	chainMgr, err := NewCaChainMgr(content)
	if err != nil {
		return fmt.Errorf("new ca Chain Mgr failed: %s", err.Error())
	}

	if err = chainMgr.CheckCertChain(chainOpts...); err != nil {
		return fmt.Errorf("check cert chain failed: %s", err.Error())
	}

	return nil
}

// CheckCertsOverdue is used to check if certs is within the allowed expiration time from PEM cert content
func CheckCertsOverdue(content []byte, overdueTime int) error {
	chainMgr, err := NewCaChainMgr(content)
	if err != nil {
		return fmt.Errorf("new ca chain mgr failed: %s", err.Error())
	}

	if err = chainMgr.CheckCertsOverdue(overdueTime); err != nil {
		return fmt.Errorf("check cert chain overdue time failed: %s", err.Error())
	}

	return nil
}

// CheckPemCertPool is used to check a cert pool from PEM cert content
func CheckPemCertPool(content []byte, chainOpts ...CaChainCheckOptions) error {
	chainMgr, err := NewCaChainMgr(content)
	if err != nil {
		return fmt.Errorf("new ca Chain Mgr failed: %s", err.Error())
	}

	if err = chainMgr.CheckCertPool(chainOpts...); err != nil {
		return fmt.Errorf("check cert chain failed: %s", err.Error())
	}

	return nil
}
