// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package appchecker  environment variable checker
package appchecker

import (
	"fmt"

	"huawei.com/mindx/common/checker"
)

// GetEnvVarChecker [method] for get environment variable checker
func GetEnvVarChecker(field string) *EnvVarChecker {
	return &EnvVarChecker{
		modelChecker: checker.ModelChecker{Field: field, Required: true},
	}
}

// EnvVarChecker [struct] for env var checker
type EnvVarChecker struct {
	modelChecker checker.ModelChecker
}

func (evc *EnvVarChecker) init() {
	evc.modelChecker.Checker = checker.GetAndChecker(
		checker.GetRegChecker("Name", envNameReg, true),
		checker.GetRegChecker("Value", envValueReg, true),
	)
}

// Check [method] for check environment variable parameters
func (evc *EnvVarChecker) Check(data interface{}) checker.CheckResult {
	evc.init()
	checkResult := evc.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("env var checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}
