// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller for setting handler checker
package edgeinstaller

import (
	"fmt"

	"huawei.com/mindxedge/base/common/checker/checker"
)

// NewSetEdgeAccountChecker [method] for getting set edge account checker struct
func NewSetEdgeAccountChecker() *setEdgeAccountChecker {
	return &setEdgeAccountChecker{}
}

type setEdgeAccountChecker struct {
	modelChecker checker.ModelChecker
}

func (sac *setEdgeAccountChecker) init() {
	sac.modelChecker.Required = true
	sac.modelChecker.Checker = checker.GetAndChecker(
		checker.GetRegChecker("Account", accountReg, true),
		checker.GetPwdChecker("Account", "Password", passwordMinLen, passwordMaxLen, true),
		checker.GetStringEqualChecker("Password", "ConfirmPassword"),
	)
}

// Check [method] for set edge account checker
func (sac *setEdgeAccountChecker) Check(data interface{}) checker.CheckResult {
	sac.init()
	checkResult := sac.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("set edge account checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}
