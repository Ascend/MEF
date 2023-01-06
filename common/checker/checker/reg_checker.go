// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package checker

import (
	"fmt"
	"regexp"

	"huawei.com/mindxedge/base/common/checker/valuer"
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
	compile := regexp.MustCompile(rc.reg)
	var matchFlag = compile.MatchString(stringValue)
	if !matchFlag {
		// todo 删除错误的详细信息
		return NewFailedResult(
			fmt.Sprintf("regex checker Check [%s] failed:the string value [%s] not match [%s]",
				rc.filed, stringValue, rc.reg))
	}
	return NewSuccessResult()
}
