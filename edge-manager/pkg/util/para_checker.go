// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package util

import (
	"fmt"
	"math"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/checker/checker"
)

// NewPaginationQueryChecker [method] for getting delete app checker struct
func NewPaginationQueryChecker() *paginationQueryChecker {
	return &paginationQueryChecker{}
}

type paginationQueryChecker struct {
	modelChecker checker.ModelChecker
}

func (pqc *paginationQueryChecker) init() {
	pqc.modelChecker.Required = true
	pqc.modelChecker.Checker = checker.GetAndChecker(
		checker.GetUintChecker("PageNum", common.DefaultPage, math.MaxInt64, true),
		checker.GetUintChecker("PageSize", common.DefaultMinPageSize, common.DefaultMaxPageSize, true),
		checker.GetRegChecker("Name", common.PaginationNameReg, false),
	)
}

func (pqc *paginationQueryChecker) Check(data interface{}) checker.CheckResult {
	pqc.init()
	checkResult := pqc.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(
			fmt.Sprintf("pagination query checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}
