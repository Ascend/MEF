// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package appchecker container port checker
package appchecker

import (
	"fmt"

	"huawei.com/mindx/common/checker"
)

// GetContainerPortChecker [method] for get container port checker
func GetContainerPortChecker(field string) *ContainerPortChecker {
	return &ContainerPortChecker{
		modelChecker: checker.ModelChecker{Field: field, Required: true},
	}
}

// ContainerPortChecker [struct] for container port checker
type ContainerPortChecker struct {
	modelChecker checker.ModelChecker
}

func (evc *ContainerPortChecker) init() {
	evc.modelChecker.Checker = checker.GetAndChecker(
		checker.GetRegChecker("Name", nameReg, true),
		checker.GetStringChoiceChecker("Proto", []string{"TCP", "UDP"}, true),
		checker.GetIntChecker("ContainerPort", minContainerPort, maxContainerPort, true),
		checker.GetIpV4Checker("HostIP", true),
		checker.GetIntChecker("HostPort", minHostPort, maxHostPort, true),
	)
}

// Check [method] for check container port parameter
func (evc *ContainerPortChecker) Check(data interface{}) checker.CheckResult {
	evc.init()
	checkResult := evc.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("container port checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}
