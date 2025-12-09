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
