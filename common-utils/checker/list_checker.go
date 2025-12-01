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
	"reflect"

	"huawei.com/mindx/common/checker/valuer"
)

// GetListChecker [method] for get list checker
func GetListChecker(field string, elemChecker checkerIntf, minLen, maxLen int, required bool) *ListChecker {
	return &ListChecker{
		field:          field,
		elementChecker: elemChecker,
		minLength:      minLen,
		maxLength:      maxLen,
		required:       required,
	}
}

// ListChecker [struct] for list checker
type ListChecker struct {
	field          string
	elementChecker checkerIntf
	minLength      int
	maxLength      int
	required       bool
	valuer         valuer.ListValuer
}

// Check [method] for do list checker
func (lc *ListChecker) Check(data interface{}) CheckResult {
	value, err := lc.valuer.GetValue(data, lc.field)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) && !lc.required {
			return NewSuccessResult()
		}
		return NewFailedResult(fmt.Sprintf("list checker get field [%s] value failed:%v", lc.field, err))
	}
	length := value.Len()
	if length < lc.minLength || length > lc.maxLength {
		return NewFailedResult(fmt.Sprintf("list checker Check len [%d] failed, not in [%d, %d]",
			length, lc.minLength, lc.maxLength))
	}

	return lc.checkElement(value)
}

func (lc *ListChecker) checkElement(listValue *reflect.Value) CheckResult {
	for i := 0; i < listValue.Len(); i++ {
		checkResult := lc.elementChecker.Check(listValue.Index(i))
		if !checkResult.Result {
			return NewFailedResult(fmt.Sprintf("list checker Check faild: %s", checkResult.Reason))
		}
	}
	return NewSuccessResult()
}
