// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package checker

import (
	"fmt"

	"huawei.com/mindx/common/checker/valuer"
)

// ExistChecker [struct] for exist checker
type ExistChecker struct {
	field  string
	valuer valuer.ExistValuer
}

// GetExistChecker [method] for get exist checker
func GetExistChecker(field string) *ExistChecker {
	return &ExistChecker{
		field: field,
	}
}

// Check [method] for do exist check
func (bc *ExistChecker) Check(data interface{}) CheckResult {
	existFlag, err := bc.valuer.GetValue(data, bc.field)
	if existFlag {
		return NewSuccessResult()
	}
	return NewFailedResult(fmt.Sprintf("exist checker Check field [%s] failed, err:%v", bc.field, err))
}
