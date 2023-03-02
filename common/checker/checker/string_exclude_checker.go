// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package checker for string exclude checker
package checker

import (
	"fmt"
	"strings"

	"huawei.com/mindxedge/base/common/checker/valuer"
)

// GetStringExcludeChecker [method] for get string exclude checker
func GetStringExcludeChecker(field string, excludeWords []string, required bool) *StringExcludeChecker {
	return &StringExcludeChecker{
		filed:        field,
		excludeWords: excludeWords,
		required:     required,
		valuer:       valuer.StringValuer{},
	}
}

// StringExcludeChecker [struct] for string choice checker
type StringExcludeChecker struct {
	filed        string
	excludeWords []string
	required     bool
	valuer       valuer.StringValuer
}

// Check [method] for do string choice check
func (scc *StringExcludeChecker) Check(data interface{}) CheckResult {
	srcString, err := scc.valuer.GetValue(data, scc.filed)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) && !scc.required {
			return NewSuccessResult()
		}
		return NewFailedResult(fmt.Sprintf("string exclude words checker get field [%s] value failed:%v",
			scc.filed, err))
	}
	for _, word := range scc.excludeWords {
		if strings.Contains(srcString, word) {
			// todo 删除错误的详细信息
			return NewFailedResult(fmt.Sprintf("string excludeWords words checker Check [%s] failed: "+
				"the value[%s], contains exclude words [%v]", scc.filed, srcString, word))
		}
	}
	return NewSuccessResult()
}
