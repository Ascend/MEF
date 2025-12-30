// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
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

// UintChecker [struct] for Uint Checker
type UintChecker struct {
	filed    string
	max      uint64
	min      uint64
	required bool
	valuer   valuer.UintValuer
}

// GetUintChecker [method] for get uint checker
func GetUintChecker(filed string, min, max uint64, required bool) *UintChecker {
	return &UintChecker{
		filed:    filed,
		min:      min,
		max:      max,
		required: required,
		valuer:   valuer.UintValuer{},
	}
}

// Check [method] for do uint check
func (uc *UintChecker) Check(data interface{}) CheckResult {
	value, err := uc.valuer.GetValue(data, uc.filed)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) && !uc.required {
			return NewSuccessResult()
		}
		return NewFailedResult(fmt.Sprintf("Uint checker get field [%s] value failed:%v", uc.filed, err))
	}
	if value < uc.min || value > uc.max {
		return NewFailedResult(
			fmt.Sprintf("Uint checker Check [%s] failed: the value[%d] not in range [%d, %d]",
				uc.filed, value, uc.min, uc.max))
	}

	return NewSuccessResult()
}
