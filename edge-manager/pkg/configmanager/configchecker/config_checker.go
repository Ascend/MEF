// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package configchecker image config checker
package configchecker

import (
	"fmt"
	"math"

	"huawei.com/mindxedge/base/common/checker/checker"
)

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
			checker.GetIpChecker("IP", true),
			checker.GetRegChecker("Domain", dnsReg, true),
		),
		checker.GetListChecker("Password",
			checker.GetUintChecker("", 0, math.MaxUint8, true),
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
