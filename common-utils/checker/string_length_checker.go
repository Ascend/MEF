// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package checker for string exclude checker
package checker

import (
	"fmt"

	"huawei.com/mindx/common/checker/valuer"
)

// GetStringLengthChecker [method] for get string length checker
func GetStringLengthChecker(field string, minLength, maxLength int, required bool) *StringLengthChecker {
	return &StringLengthChecker{
		filed:     field,
		minLength: minLength,
		maxLength: maxLength,
		required:  required,
		valuer:    valuer.StringValuer{},
	}
}

// StringLengthChecker [struct] for string length checker
type StringLengthChecker struct {
	filed     string
	minLength int
	maxLength int
	required  bool
	valuer    valuer.StringValuer
}

// Check [method] for do string choice check
func (slc *StringLengthChecker) Check(data interface{}) CheckResult {
	srcString, err := slc.valuer.GetValue(data, slc.filed)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) && !slc.required {
			return NewSuccessResult()
		}
		return NewFailedResult(fmt.Sprintf("string length checker get field [%s] value failed:%v",
			slc.filed, err))
	}
	if len(srcString) > slc.maxLength || len(srcString) < slc.minLength {
		return NewFailedResult(fmt.Sprintf("string length checker Check [%s] failed: "+
			"the length not in range from %d to %d", slc.filed, slc.minLength, slc.maxLength))
	}
	return NewSuccessResult()
}
