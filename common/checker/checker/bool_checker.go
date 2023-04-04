// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package checker

import (
	"fmt"

	"huawei.com/mindxedge/base/common/checker/valuer"
)

// GetBoolChecker get bool checker
func GetBoolChecker(filed string, required bool) *BoolChecker {
	return &BoolChecker{
		field:    filed,
		required: required,
		valuer:   valuer.BoolValuer{},
	}
}

// BoolChecker [struct] for bool checker
type BoolChecker struct {
	field    string
	required bool
	valuer   valuer.BoolValuer
}

// Check method
func (bc *BoolChecker) Check(data interface{}) CheckResult {
	_, err := bc.valuer.GetValue(data, bc.field)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) && !bc.required {
			return NewSuccessResult()
		}
		return NewFailedResult(fmt.Sprintf("Bool checker get field[%s] value failed:%v", bc.field, err))
	}
	return NewSuccessResult()
}
