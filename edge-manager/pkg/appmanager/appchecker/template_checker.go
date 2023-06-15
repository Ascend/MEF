// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package appchecker

import (
	"fmt"

	"huawei.com/mindx/common/checker"
)

// NewCreateTemplateChecker to get checker [struct] for create template check
func NewCreateTemplateChecker() *createTemplateChecker {
	return &createTemplateChecker{}
}

// NewUpdateTemplateChecker to get checker [struct] for update template check
func NewUpdateTemplateChecker() *updateTemplateChecker {
	return &updateTemplateChecker{}
}

type createTemplateChecker struct {
	modelChecker checker.ModelChecker
}

type updateTemplateChecker struct {
	modelChecker checker.ModelChecker
	createTemplateChecker
}

func (ctc *createTemplateChecker) init() {
	ctc.modelChecker.Required = true
	ctc.modelChecker.Checker = checker.GetAndChecker(
		checker.GetRegChecker("Name", nameReg, true),
		checker.GetRegChecker("Description", descriptionReg, false),
		checker.GetListChecker(
			"Containers",
			GetContainerChecker(""),
			minContainerCountInPod,
			maxContainerCountInPod,
			true,
		),
	)
}

func (utc *updateTemplateChecker) init() {
	utc.modelChecker.Required = true
	utc.createTemplateChecker.init()
	utc.modelChecker.Checker = checker.GetAndChecker(
		checker.GetUintChecker("Id", minTemplateId, maxTemplateId, true),
		&utc.createTemplateChecker.modelChecker,
	)
}

// Check [method] for create template check
func (ctc *createTemplateChecker) Check(data interface{}) checker.CheckResult {
	ctc.init()
	checkResult := ctc.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("create template checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

// Check [method] for update template check
func (utc *updateTemplateChecker) Check(data interface{}) checker.CheckResult {
	utc.init()
	checkResult := utc.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("update template checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}
