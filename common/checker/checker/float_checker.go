// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package checker

import (
	"fmt"

	"huawei.com/mindxedge/base/common/checker/valuer"
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
		// todo 删除错误信息的返回
		return NewFailedResult(fmt.Sprintf("Float checker get field[%s] value failed:%v", fc.field, err))
	}
	if value < fc.min || value > fc.max {
		return NewFailedResult(
			fmt.Sprintf("Float checker Check [%s] failed: the value[%f] not in range [%f, %f]",
				fc.field, value, fc.min, fc.max))
	}

	return NewSuccessResult()
}
