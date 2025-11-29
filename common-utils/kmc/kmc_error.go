// Copyright(C) Huawei Technologies Co.,Ltd. 2022-2022. All rights reserved.

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
