// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package checker

import (
	"fmt"
	"reflect"

	"huawei.com/mindxedge/base/common/checker/valuer"
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
		// todo 删除错误信息的返回
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
