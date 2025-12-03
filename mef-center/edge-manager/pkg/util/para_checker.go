// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package util

import (
	"fmt"
	"math"

	"huawei.com/mindx/common/checker"

	"huawei.com/mindxedge/base/common"
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
		checker.GetUintChecker("PageNum", common.DefaultPage, math.MaxInt32, true),
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
