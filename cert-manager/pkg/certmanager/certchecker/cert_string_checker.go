// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package certchecker cert string checker
package certchecker

import (
	"fmt"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/checker/valuer"
)

// GetStringChecker [method] for get string checker
func GetStringChecker(field string, f func(string) error, required bool) *StringChecker {
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
	f        func(string) error
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
	if err := sc.f(targetString); err != nil {
		return checker.NewFailedResult(fmt.Sprintf("string checker Check [%s] failed: %v", sc.filed, err))
	}
	return checker.NewSuccessResult()

}
