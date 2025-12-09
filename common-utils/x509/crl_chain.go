// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package x509

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	"huawei.com/mindx/common/fileutils"
)

const crlBegin = "-----BEGIN X509 CRL-----"

// CrlChainMgr is the struct to manager a crl chain
type CrlChainMgr struct {
	crls []*pkix.CertificateList
}

// NewCrlMgr is used to create and init a NewCrlMgr, the content support PEM format crl
func NewCrlMgr(content []byte) (*CrlChainMgr, error) {
	crlMgr := CrlChainMgr{}

	err := crlMgr.parsePemCrl(content)
	if err != nil {
		return nil, err
	}

	return &crlMgr, nil
}

// GetCrls get crls
func (ccm *CrlChainMgr) GetCrls() []*pkix.CertificateList {
	return ccm.crls
}

func (ccm *CrlChainMgr) parsePemCrl(content []byte) error {
	if len(content) <= 0 {
		return errors.New("content is empty")
	}

	var block *pem.Block
	var derCrl []byte
	certs := strings.Split(string(content), crlBegin)
	if len(certs) <= 1 {
		return fmt.Errorf("no PEM crl contains")
	}

	for _, cert := range certs[1:] {
		sigContent := []byte(crlBegin + cert)
		block, _ = pem.Decode(sigContent)
		if block == nil {
			return errors.New("parse crl failed: no content decoded")
		}
		if block.Type != "X509 CRL" || len(block.Headers) != 0 {
			return errors.New("invalid crl bytes")
		}
		derCrl = block.Bytes

		crl, err := x509.ParseCRL(derCrl)
		if err != nil {
			return fmt.Errorf("parse single crl failed: %s", err.Error())
		}

		ccm.crls = append(ccm.crls, crl)
		if len(ccm.crls) > maxCertCount {
			return fmt.Errorf("the crls count exceeds limitation")
		}
	}

	return nil
}

// CheckCrl func is the main func to check the importing crl and related ca certs
func (ccm *CrlChainMgr) CheckCrl(certData CertData) error {
	if err := ccm.checkCrlValid(); err != nil {
		return err
	}
	certBytes, err := certData.GetCertBytes()
	if err != nil {
		return err
	}
	chainMgr, err := NewCaChainMgr(certBytes)
	if err != nil {
		return fmt.Errorf("init ca chain manager instance failed: %v", err)
	}
	if err := ccm.compareWithCerts(chainMgr.GetCerts()); err != nil {
		return ErrCrlCertNotMatch
	}

	return nil
}

func (ccm *CrlChainMgr) checkCrlValid() error {
	for _, crl := range ccm.crls {
		issueDate := crl.TBSCertList.ThisUpdate
		nextUpdateDate := crl.TBSCertList.NextUpdate
		timeNow := time.Now().In(time.UTC)
		if timeNow.After(nextUpdateDate) || timeNow.Before(issueDate) {
			return ErrCrlInvalidUpdateTime
		}
	}
	return nil
}

func (ccm *CrlChainMgr) compareWithCerts(certs []*x509.Certificate) error {
	if len(certs) != len(ccm.crls) {
		return fmt.Errorf("the importing certs count %d does not equal the crls count %d",
			len(certs), len(ccm.crls))
	}
	for _, crl := range ccm.crls {
		verified := false
		var certsDup []*x509.Certificate
		for idx, cert := range certs {
			if err := cert.CheckCRLSignature(crl); err != nil {
				continue
			}
			verified = true
			certsDup = append(certs[:idx], certs[idx+1:]...)
			break
		}
		certs = certsDup

		if !verified {
			return errors.New("the crl chain contains unverified crl")
		}
	}

	return nil
}

// CheckCrlsChainReturnContent is used to check a crl chain file and returns its content
func CheckCrlsChainReturnContent(crlPath, certPath string) ([]byte, error) {
	crlContent, err := fileutils.LoadFile(crlPath)
	if err != nil {
		return nil, err
	}

	chainMgr, err := NewCrlMgr(crlContent)
	if err != nil {
		return nil, fmt.Errorf("new crl Chain Mgr failed: %s", err.Error())
	}

	if err = chainMgr.CheckCrl(CertData{CertPath: certPath}); err != nil {
		return nil, fmt.Errorf("check crl chain failed: %s", err.Error())
	}

	return crlContent, nil
}

// ParseCrls is used to parse crls from a crl path or []byte content
func ParseCrls(crlData *CrlData) ([]*pkix.CertificateList, error) {
	crlContent, err := crlData.GetCrlBytes()
	if err != nil {
		return nil, err
	}
	chainMgr, err := NewCrlMgr(crlContent)
	if err != nil {
		return nil, fmt.Errorf("new crl Chain Mgr failed: %s", err.Error())
	}

	return chainMgr.GetCrls(), nil
}
