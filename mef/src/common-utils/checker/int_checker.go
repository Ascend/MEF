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

// IntChecker [struct] for int checker
type IntChecker struct {
	field    string
	max      int64
	min      int64
	required bool
	valuer   valuer.IntValuer
}

// GetIntChecker [method] for get integer checker
func GetIntChecker(filed string, min, max int64, required bool) *IntChecker {
	return &IntChecker{
		field:    filed,
		min:      min,
		max:      max,
		required: required,
		valuer:   valuer.IntValuer{},
	}
}

// Check [method] for do int check
func (ic *IntChecker) Check(data interface{}) CheckResult {
	value, err := ic.valuer.GetValue(data, ic.field)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) && !ic.required {
			return NewSuccessResult()
		}
		return NewFailedResult(fmt.Sprintf("Int checker get field [%s] value failed:%v", ic.field, err))
	}
	if value < ic.min || value > ic.max {
		return NewFailedResult(fmt.Sprintf("Int checker Check [%s] failed: the value[%d] not in range [%d, %d]",
			ic.field, value, ic.min, ic.max))
	}

	return NewSuccessResult()
}
