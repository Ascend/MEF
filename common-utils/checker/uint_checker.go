// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

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
