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
	"regexp"

	"huawei.com/mindx/common/checker/valuer"
)

// GetRegChecker [method] for get regex checker
func GetRegChecker(filed, reg string, required bool) *RegChecker {
	return &RegChecker{
		filed:    filed,
		reg:      reg,
		required: required,
		valuer:   valuer.StringValuer{},
	}
}

// RegChecker [struct] for regex checker
type RegChecker struct {
	filed    string
	reg      string
	required bool
	valuer   valuer.StringValuer
}

// Check [method] for do regex check
func (rc *RegChecker) Check(data interface{}) CheckResult {
	stringValue, err := rc.valuer.GetValue(data, rc.filed)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) && !rc.required {
			return NewSuccessResult()
		}
		return NewFailedResult(fmt.Sprintf("regex checker get field[%s] value failed:%v", rc.filed, err))
	}
	compile, err := regexp.Compile(rc.reg)
	if err != nil {
		return NewFailedResult("regex checker compile reg failed")
	}
	var matchFlag = compile.MatchString(stringValue)
	if !matchFlag {
		return NewFailedResult(
			fmt.Sprintf("regex checker Check [%s] failed:the string value not match requirement",
				rc.filed))
	}
	return NewSuccessResult()
}
