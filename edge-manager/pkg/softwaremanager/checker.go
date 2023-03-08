// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package softwaremanager para checker
package softwaremanager

import (
	"fmt"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/checker/checker"
)

type sfwAuthInfoChecker struct {
	sfwAuthInfoChecker checker.ModelChecker
}

func newSftAuthInfoChecker() *sfwAuthInfoChecker {
	return &sfwAuthInfoChecker{}
}

func (cc *sfwAuthInfoChecker) init() {
	cc.sfwAuthInfoChecker.Checker = checker.GetAndChecker(
		checker.GetRegChecker("UserName", "^[a-zA-Z0-9]{6,32}$", true),
		checker.GetExistChecker("Password"),
	)
}

func (cc *sfwAuthInfoChecker) Check(data interface{}) checker.CheckResult {
	cc.init()
	checkResult := cc.sfwAuthInfoChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("sfw auth info check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

type sfwUrlInfoChecker struct {
	sftAuthInfoChecker checker.ModelChecker
}

func newSfwUrlInfoChecker() *sfwUrlInfoChecker {
	return &sfwUrlInfoChecker{}
}

func (cc *sfwUrlInfoChecker) init() {
	cc.sftAuthInfoChecker.Checker = checker.GetAndChecker(
		checker.GetStringChoiceChecker("Option",
			[]string{opAdd, opDelete, opSync}, true),
		checker.GetListChecker("UrlInfos", getUrlInfoChecker("", true),
			1, maxSftUrlCount, true),
	)
}

func (cc *sfwUrlInfoChecker) Check(data interface{}) checker.CheckResult {
	cc.init()
	checkResult := cc.sftAuthInfoChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("sft url info check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

type urlInfoChecker struct {
	modelChecker checker.ModelChecker
}

func (d *urlInfoChecker) init() {
	d.modelChecker.Checker = checker.GetAndChecker(
		checker.GetStringChoiceChecker("Type",
			[]string{common.EdgeCore, common.DevicePlugin}, true),
		checker.GetHttpsUrlChecker("Url", true),
		checker.GetRegChecker("Version", "^[a-zA-Z0-9-_.]{1,32}$", true),
		checker.GetRegChecker("CreatedAt", "^[0-9: -]{1,64}$", true),
	)
}

func (d *urlInfoChecker) Check(data interface{}) checker.CheckResult {
	d.init()

	checkResult := d.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("url info check failed: %s", checkResult.Reason))
	}

	return checker.NewSuccessResult()
}

func getUrlInfoChecker(field string, required bool) *urlInfoChecker {
	return &urlInfoChecker{
		modelChecker: checker.ModelChecker{Field: field, Required: required},
	}
}
