// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// kmc interface
package kmc

import "fmt"

// KeKmcError kmc error code and error
type KeKmcError struct {
	code int
	err  string
}

const (
	kmcLibNotFound   = 21 // not init
	kmcNotInit       = 22 // not init
	invalidDomainID  = 23 // invalid domain ID
	paramCheckFailed = 24 // parameter check failed
)

var (
	kmcLibErr     = NewKmcError(kmcLibNotFound, "cannot found kmc shared library")
	kmcNotInitErr = NewKmcError(kmcNotInit, "not init")
	domainIDErr   = NewKmcError(invalidDomainID, "invalid domainID")
	paramCheckErr = NewKmcError(paramCheckFailed, "parameter check failed")
)

// NewKmcError create a kmc error object
func NewKmcError(code int, err string) *KeKmcError {
	return &KeKmcError{code: code, err: err}
}

// Code error code
func (e *KeKmcError) Code() int {
	return e.code
}

// Error error information
func (e *KeKmcError) Error() string {
	return fmt.Sprintf("%d: %s", e.code, e.err)
}
