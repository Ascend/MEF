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
	"math"

	"huawei.com/mindx/common/checker/valuer"
)

// FloatChecker [struct] for float checker
type FloatChecker struct {
	field    string
	max      float64
	min      float64
	required bool
	valuer   valuer.FloatValuer
}

// GetFloatChecker [method] for get float checker
func GetFloatChecker(filed string, min, max float64, required bool) *FloatChecker {
	return &FloatChecker{
		field:    filed,
		min:      min,
		max:      max,
		required: required,
		valuer:   valuer.FloatValuer{},
	}
}

// Check [method] do float check
func (fc *FloatChecker) Check(data interface{}) CheckResult {
	value, err := fc.valuer.GetValue(data, fc.field)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) && !fc.required {
			return NewSuccessResult()
		}
		return NewFailedResult(fmt.Sprintf("Float checker get field[%s] value failed:%v", fc.field, err))
	}

	// Handling the NaN situation
	if math.IsNaN(value) {
		return NewFailedResult(fmt.Sprintf("Float checker check [%s] failed: NaN is not allowed", fc.field))
	}

	if value < fc.min || value > fc.max {
		return NewFailedResult(
			fmt.Sprintf("Float checker Check [%s] failed: the value[%f] not in range [%f, %f]",
				fc.field, value, fc.min, fc.max))
	}

	return NewSuccessResult()
}
