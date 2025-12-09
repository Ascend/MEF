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

import "fmt"

// CertError - wrap cert and crl related errors. For identify different kind errors
type CertError struct {
	ErrCode int
	ErrDesc string
}

// Error - implement error interface
func (e *CertError) Error() string {
	return fmt.Sprintf("x509 error. code: [%v], reason: [%v]", e.ErrCode, e.ErrDesc)
}

// Define error types when required
var (
	ErrCertParseFailed      = &CertError{ErrCode: 1001, ErrDesc: "parse x509 certificate failed"}
	ErrCertExpired          = &CertError{ErrCode: 1002, ErrDesc: "certificate is already expired"}
	ErrCrlParseFailed       = &CertError{ErrCode: 1003, ErrDesc: "parse x509 CRL failed"}
	ErrCrlExpired           = &CertError{ErrCode: 1004, ErrDesc: "CRL is already expired"}
	ErrCrlCertNotMatch      = &CertError{ErrCode: 1005, ErrDesc: "CRLs and certificates doesn't match"}
	ErrCrlInvalidUpdateTime = &CertError{ErrCode: 1006,
		ErrDesc: "CRL time before the issuing date or after the nextupdate date"}
)
