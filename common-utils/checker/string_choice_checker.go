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
	"sort"

	"huawei.com/mindx/common/checker/valuer"
)

// GetStringChoiceChecker [method] for get string choice checker
func GetStringChoiceChecker(field string, choices []string, required bool) *StringChoiceChecker {
	return &StringChoiceChecker{
		filed:    field,
		choices:  choices,
		required: required,
		valuer:   valuer.StringValuer{},
	}
}

// StringChoiceChecker [struct] for string choice checker
type StringChoiceChecker struct {
	filed    string
	choices  []string
	required bool
	valuer   valuer.StringValuer
}

// Check [method] for do string choice check
func (scc *StringChoiceChecker) Check(data interface{}) CheckResult {
	targetString, err := scc.valuer.GetValue(data, scc.filed)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) && !scc.required {
			return NewSuccessResult()
		}
		return NewFailedResult(fmt.Sprintf("string choice checker get field [%s] value failed:%v", scc.filed, err))
	}
	sort.Strings(scc.choices)
	index := sort.SearchStrings(scc.choices, targetString)
	if index < len(scc.choices) && scc.choices[index] == targetString {
		return NewSuccessResult()
	}
	return NewFailedResult(fmt.Sprintf("string choice checker Check [%s] failed: the value not in %v",
		scc.filed, scc.choices))
}
