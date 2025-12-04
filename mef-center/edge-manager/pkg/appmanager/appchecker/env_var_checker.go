// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
