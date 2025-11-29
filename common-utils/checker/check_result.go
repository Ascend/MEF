// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package checker

import "fmt"

// CheckResult [struct] for check result
type CheckResult struct {
	// Result [struct field] for result flag
	Result bool
	// Reason [struct field] for result reason
	Reason string
}

// String [method] for print result
func (cr *CheckResult) String() string {
	return fmt.Sprintf("Result=%t, Reason=%s\n", cr.Result, cr.Reason)
}

// NewSuccessResult [method] for generate a new success CheckResult
func NewSuccessResult() CheckResult {
	return CheckResult{Result: true, Reason: ""}
}

// NewFailedResult [method] for generate failed CheckResult
func NewFailedResult(reason string) CheckResult {
	return CheckResult{Result: false, Reason: reason}
}
