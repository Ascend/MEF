// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package appchecker create app checker
package appchecker

import (
	"fmt"

	"huawei.com/mindxedge/base/common/checker/checker"
)

// CreateAppChecker [struct] for create app check
type CreateAppChecker struct {
	modelChecker checker.ModelChecker
}

func (cac *CreateAppChecker) init() {
	cac.modelChecker.Required = true
	cac.modelChecker.Checker = checker.GetAndChecker(
		checker.GetRegChecker("AppName", nameReg, true),
		checker.GetRegChecker("Description", descriptReg, true),
		checker.GetListChecker(
			"Containers",
			GetContainerChecker(""),
			minContainerCountInPod,
			maxContainerCountInPod,
			true,
		),
	)
}

// Check [method] for check create app checker
func (cac *CreateAppChecker) Check(data interface{}) checker.CheckResult {
	cac.init()
	checkResult := cac.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("create app checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}
