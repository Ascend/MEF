// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package usermgr for
package usermgr

import (
	"fmt"
	"huawei.com/mindxedge/base/common/checker/checker"
)

const nameReg = "[a-z0-9-_]{0,10}"

// newLoginChecker [method] for getting login checker struct
func newLoginChecker() *loginChecker {
	return &loginChecker{}
}

func newFirstChangeChecker() *firstChangeChecker {
	return &firstChangeChecker{}
}

func newChangeChecker() *changeChecker {
	return &changeChecker{}
}

func newLockChecker() *lockChecker {
	return &lockChecker{}
}

type loginChecker struct {
	modelChecker checker.ModelChecker
}

type firstChangeChecker struct {
	modelChecker checker.ModelChecker
}

type changeChecker struct {
	modelChecker checker.ModelChecker
}

type lockChecker struct {
	modelChecker checker.ModelChecker
}

func (cac *loginChecker) init() {
	cac.modelChecker.Required = true
	cac.modelChecker.Checker = checker.GetAndChecker(
		checker.GetRegChecker("Username", nameReg, true),
		checker.GetExistChecker("Password"),
		checker.GetIpChecker("Ip", true),
	)
}

func (cac *firstChangeChecker) init() {
	cac.modelChecker.Required = true
	cac.modelChecker.Checker = checker.GetAndChecker(
		checker.GetRegChecker("Username", nameReg, true),
		checker.GetExistChecker("Password"),
		checker.GetExistChecker("RePassword"),
		checker.GetIpChecker("Ip", true),
	)
}

func (cac *changeChecker) init() {
	cac.modelChecker.Required = true
	cac.modelChecker.Checker = checker.GetAndChecker(
		checker.GetRegChecker("Username", nameReg, true),
		checker.GetExistChecker("Password"),
		checker.GetExistChecker("OldPassword"),
		checker.GetExistChecker("RePassword"),
		checker.GetIpChecker("Ip", true),
	)
}

func (cac *lockChecker) init() {
	cac.modelChecker.Required = true
	cac.modelChecker.Checker = checker.GetAndChecker(
		checker.GetIpChecker("TargetIp", true),
		checker.GetIpChecker("Ip", true),
	)
}

func (cac *loginChecker) Check(data interface{}) checker.CheckResult {
	cac.init()
	checkResult := cac.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("login checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

func (cac *firstChangeChecker) Check(data interface{}) checker.CheckResult {
	cac.init()
	checkResult := cac.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("change pwd checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

func (cac *changeChecker) Check(data interface{}) checker.CheckResult {
	cac.init()
	checkResult := cac.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("change pwd checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

func (cac *lockChecker) Check(data interface{}) checker.CheckResult {
	cac.init()
	checkResult := cac.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("change pwd checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}
