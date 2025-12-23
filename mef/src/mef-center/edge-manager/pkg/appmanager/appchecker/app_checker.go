// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package appchecker create app checker
package appchecker

import (
	"fmt"

	"huawei.com/mindx/common/checker"
)

// NewCreateAppChecker [method] for getting create app checker struct
func NewCreateAppChecker() *createAppChecker {
	return &createAppChecker{}
}

// IdChecker [method] for getting query app checker struct
func IdChecker() *checker.UintChecker {
	return checker.GetUintChecker("", minAppId, maxAppId, true)
}

// NewUpdateAppChecker [method] for getting update app checker struct
func NewUpdateAppChecker() *updateAppChecker {
	return &updateAppChecker{}
}

// NewDeleteAppChecker [method] for getting delete app checker struct
func NewDeleteAppChecker() *deleteAppChecker {
	return &deleteAppChecker{}
}

// NewDeployAppChecker [method] for getting delete app checker struct
func NewDeployAppChecker() *deployAppChecker {
	return &deployAppChecker{}
}

// NewUndeployAppChecker [method] for getting delete app checker struct
func NewUndeployAppChecker() *undeployAppChecker {
	return &undeployAppChecker{}
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

type deployAppChecker struct {
	modelChecker checker.ModelChecker
}

type undeployAppChecker struct {
	deployAppChecker
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

func (dac *deleteAppChecker) init() {
	dac.idListChecker = *checker.GetUniqueListChecker(
		"AppIDs",
		checker.GetUintChecker("", minAppId, maxAppId, true),
		minList,
		maxList,
		true)
}

func (dac *deployAppChecker) init() {
	dac.modelChecker.Required = true
	dac.modelChecker.Checker = checker.GetAndChecker(
		checker.GetUintChecker("AppID", minAppId, maxAppId, true),
		checker.GetUniqueListChecker(
			"NodeGroupIds",
			checker.GetUintChecker("", minNodeGroupId, maxNodeGroupId, true),
			minList,
			maxList,
			true),
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

func (dac *deleteAppChecker) Check(data interface{}) checker.CheckResult {
	dac.init()
	checkResult := dac.idListChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("delete app checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

func (dac *deployAppChecker) Check(data interface{}) checker.CheckResult {
	dac.init()
	checkResult := dac.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("deploy app checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

func (uac *undeployAppChecker) Check(data interface{}) checker.CheckResult {
	uac.init()
	checkResult := uac.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("undeploy app checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}
