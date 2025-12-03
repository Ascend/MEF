// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
