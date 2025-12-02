// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package x509 provides the capability of x509.
package x509

import (
	"errors"
	"time"

	"huawei.com/mindx/common/fileutils"
)

// CertStatus  the certificate valid period
type CertStatus struct {
	NotBefore         time.Time `json:"not_before"`
	NotAfter          time.Time `json:"not_after"`
	IsCA              bool      `json:"is_ca"`
	FingerprintSHA256 string    `json:"fingerprint_sha256,omitempty"`
}

// CertData - two ways to get certificate data: fs path and memory bytes.
type CertData struct {
	CertPath    string
	CertContent []byte
}

// CrlData - two ways to get CRL data: fs path and memory bytes.
type CrlData struct {
	CrlContent []byte
	CrlPath    string
}

// GetCertBytes - load certificate data from file or memory.
// If both file path and memory content are provided, use memory content first.
func (cd *CertData) GetCertBytes() ([]byte, error) {
	if cd == nil {
		return nil, errors.New("cert data instance is invaid")
	}
	var err error
	certContent := cd.CertContent
	if len(certContent) == 0 {
		certContent, err = fileutils.LoadFile(cd.CertPath)
		if err != nil {
			return nil, err
		}
	}
	if len(certContent) == 0 {
		return nil, errors.New("empty certificate data")
	}
	return certContent, nil
}

// GetCrlBytes - load CRL data from file or memory.
// If both file path and memory content are provided, use memory content first.
func (crl *CrlData) GetCrlBytes() ([]byte, error) {
	if crl == nil {
		return nil, errors.New("crl data instance is invaid")
	}
	var err error
	crlContent := crl.CrlContent
	if len(crlContent) == 0 {
		crlContent, err = fileutils.LoadFile(crl.CrlPath)
		if err != nil {
			return nil, err
		}
	}
	if len(crlContent) == 0 {
		return nil, errors.New("empty CRL data")
	}
	return crlContent, nil
}

// ContainsData - because CRL is optional, it can be empty, we need a fast way to check if it's empty
func (crl *CrlData) ContainsData() bool {
	return len(crl.CrlContent) > 0 || fileutils.IsExist(crl.CrlPath)
}
