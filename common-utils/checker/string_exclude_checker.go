// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package checker for string exclude checker
package checker

import (
	"fmt"
	"strings"

	"huawei.com/mindx/common/checker/valuer"
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
			return NewFailedResult(fmt.Sprintf("string excludeWords words checker Check [%s] failed: "+
				"the value contains exclude words [%v]", scc.filed, word))
		}
	}
	return NewSuccessResult()
}
