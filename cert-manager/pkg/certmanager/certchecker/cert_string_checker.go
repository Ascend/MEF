// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package certchecker cert string checker
package certchecker

import (
	"fmt"

	"huawei.com/mindxedge/base/common/checker/checker"
	"huawei.com/mindxedge/base/common/checker/valuer"
)

// GetStringChecker [method] for get string checker
func GetStringChecker(field string, f func(string) bool, required bool) *StringChecker {
	return &StringChecker{
		filed:    field,
		f:        f,
		required: required,
		valuer:   valuer.StringValuer{},
	}
}

// StringChecker [struct] for string checker
type StringChecker struct {
	filed    string
	f        func(string) bool
	required bool
	valuer   valuer.StringValuer
}

// Check [method] for do string check
func (sc *StringChecker) Check(data interface{}) checker.CheckResult {
	targetString, err := sc.valuer.GetValue(data, sc.filed)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) && !sc.required {
			return checker.NewSuccessResult()
		}
		return checker.NewFailedResult(fmt.Sprintf("string checker get field [%s] value failed:%v", sc.filed, err))
	}
	if !sc.f(targetString) {
		return checker.NewFailedResult(fmt.Sprintf("string checker Check [%s] failed: the value[%s]",
			sc.filed, targetString))
	}
	return checker.NewSuccessResult()

}
