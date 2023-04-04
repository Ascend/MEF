// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package checker

import (
	"fmt"
	"sort"

	"huawei.com/mindxedge/base/common/checker/valuer"
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
	return NewFailedResult(fmt.Sprintf("string choice checker Check [%s] failed: the value[%s], not in %v",
		scc.filed, targetString, scc.choices))
}
