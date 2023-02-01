// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package appchecker create app checker
package appchecker

import (
	"fmt"
	"math"

	"huawei.com/mindxedge/base/common/checker/checker"
)

// NewCreateAppChecker [method] for getting create app checker struct
func NewCreateAppChecker() *createAppChecker {
	return &createAppChecker{}
}

// NewUpdateAppChecker [method] for getting update app checker struct
func NewUpdateAppChecker() *updateAppChecker {
	return &updateAppChecker{}
}

// NewDeleteAppChecker [method] for getting delete app checker struct
func NewDeleteAppChecker() *deleteAppChecker {
	return &deleteAppChecker{}
}

type createAppChecker struct {
	modelChecker checker.ModelChecker
}

type updateAppChecker struct {
	modelChecker checker.ModelChecker
	createAppChecker
}

type deleteAppChecker struct {
	idListChecker checker.UniqueListChecker
}

func (cac *createAppChecker) init() {
	cac.modelChecker.Required = true
	cac.modelChecker.Checker = checker.GetAndChecker(
		checker.GetRegChecker("AppName", nameReg, true),
		checker.GetRegChecker("Description", descriptionReg, true),
		checker.GetListChecker(
			"Containers",
			GetContainerChecker(""),
			minContainerCountInPod,
			maxContainerCountInPod,
			true,
		),
	)
}

func (uac *updateAppChecker) init() {
	uac.modelChecker.Required = true
	uac.createAppChecker.init()
	uac.modelChecker.Checker = checker.GetAndChecker(
		checker.GetUintChecker("AppID", minAppId, maxAppId, true),
		&uac.createAppChecker.modelChecker,
	)
}

// Check [method] for create app checker
func (cac *createAppChecker) Check(data interface{}) checker.CheckResult {
	cac.init()
	checkResult := cac.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("create app checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

// Check [method] for update app checker
func (uac *updateAppChecker) Check(data interface{}) checker.CheckResult {
	uac.init()
	checkResult := uac.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("update app checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

func (dac deleteAppChecker) Check(data interface{}) checker.CheckResult {
	listChecker := checker.GetUniqueListChecker(
		"AppIDs",
		checker.GetUintChecker("", 1, math.MaxInt64, true),
		minList,
		maxList,
		true)
	checkResult := listChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("delete app checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}
