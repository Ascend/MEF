// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

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
