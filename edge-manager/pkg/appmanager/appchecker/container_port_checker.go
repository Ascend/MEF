// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
		checker.GetStringChoiceChecker("Proto", []string{"TCP", "UDP", "SCTP"}, true),
		checker.GetIntChecker("ContainerPort", minContainerPort, maxContainerPort, true),
		checker.GetIpChecker("HostIP", true),
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
