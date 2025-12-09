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

// GetOrChecker [method] for get or checker
func GetOrChecker(checks ...checkerIntf) *OrChecker {
	return &OrChecker{
		checkers: checks,
	}
}

// OrChecker [struct] for or checker
type OrChecker struct {
	checkers []checkerIntf
}

// Check [method] for do or check
func (ac *OrChecker) Check(data interface{}) CheckResult {
	var reason string
	for _, checker := range ac.checkers {
		checkResult := checker.Check(data)
		if checkResult.Result {
			return NewSuccessResult()
		}
		reason += checkResult.Reason + ";"
	}
	return NewFailedResult(fmt.Sprintf("Or checker Check failed: %s", reason))
}
