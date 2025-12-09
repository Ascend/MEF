// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
