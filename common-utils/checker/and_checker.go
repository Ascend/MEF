// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

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
