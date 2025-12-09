// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package checker
package checker

// AndChecker [struct] for and checker
type AndChecker struct {
	checkers []checkerIntf
}

// GetAndChecker [method] for
func GetAndChecker(checks ...checkerIntf) *AndChecker {
	return &AndChecker{
		checkers: checks,
	}
}

// Check [method] for do and check
func (ac *AndChecker) Check(data interface{}) CheckResult {
	for _, checker := range ac.checkers {
		result := checker.Check(data)
		if !result.Result {
			return NewFailedResult(result.Reason)
		}
	}
	return NewSuccessResult()
}
