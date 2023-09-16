// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package configchecker image config checker
package configchecker

import (
	"fmt"
	"math"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/checker/valuer"
)

const invalidChar = ':'

type configChecker struct {
	configCheck checker.ModelChecker
}

// NewConfigChecker [method] for getting image config checker struct
func NewConfigChecker() *configChecker {
	return &configChecker{}
}

func (cc *configChecker) init() {
	cc.configCheck.Checker = checker.GetAndChecker(
		checker.GetRegChecker("Account", nameReg, true),
		checker.GetOrChecker(
			checker.GetRegChecker("Domain", domainReg, true),
			checker.GetAndChecker(
				checker.GetStringChoiceChecker("Domain", []string{""}, true),
				checker.GetIpV4Checker("IP", true),
			),
		),
		checker.GetListChecker("Password",
			checker.GetAndChecker(
				checker.GetUintChecker("", 0, math.MaxUint8, true),
				&invalidCharChecker{},
			),
			minPwdCount,
			maxPwdCount,
			true,
		),
		checker.GetIntChecker("Port", minHostPort, maxHostPort, true),
	)
}

func (cc *configChecker) Check(data interface{}) checker.CheckResult {
	cc.init()
	checkResult := cc.configCheck.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("image config checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

type invalidCharChecker struct{}

func (i *invalidCharChecker) Check(data interface{}) checker.CheckResult {
	uintValuer := valuer.UintValuer{}
	value, err := uintValuer.GetValue(data, "")
	if err != nil {
		return checker.NewFailedResult(fmt.Sprintf("get uint value failed, error: %v", err))
	}
	if value == invalidChar {
		return checker.NewFailedResult("password contains invalid character")
	}
	return checker.NewSuccessResult()
}
