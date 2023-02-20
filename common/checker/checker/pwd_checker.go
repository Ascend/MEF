// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package checker for pwd
package checker

import (
	"fmt"
	"regexp"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/checker/valuer"
	"huawei.com/mindxedge/base/common/passutils"
)

// PwdChecker [struct] for password checker
type PwdChecker struct {
	field    string
	maxLen   int
	minLen   int
	required bool
	valuer   valuer.StringValuer
}

// GetPwdChecker [method] for get password checker
func GetPwdChecker(filed string, min, max int, required bool) *PwdChecker {
	return &PwdChecker{
		field:    filed,
		minLen:   min,
		maxLen:   max,
		required: required,
		valuer:   valuer.StringValuer{},
	}
}

// Check [method] for do password check
func (pc *PwdChecker) Check(data interface{}) CheckResult {
	value, err := pc.valuer.GetValue(data, pc.field)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) && !pc.required {
			return NewSuccessResult()
		}
		// todo 删除错误信息的返回
		hwlog.RunLog.Errorf("Pwd checker get field [%s] value failed, error: %v", pc.field, err)
		return NewFailedResult(fmt.Sprintf("get field [%s] value failed", pc.field))
	}

	return pc.isPwdValid(value)
}

func (pc *PwdChecker) isPwdValid(data string) CheckResult {
	if len(data) < pc.minLen || len(data) > pc.maxLen {
		hwlog.RunLog.Errorf("Pwd checker Check [%s] failed: the length is not in range [%d, %d]",
			pc.field, pc.minLen, pc.maxLen)
		return NewFailedResult(fmt.Sprintf("the length is not in range [%d, %d]", pc.minLen, pc.maxLen))
	}

	if matched, err := regexp.MatchString(common.PassWordRegex, data); err != nil || !matched {
		hwlog.RunLog.Errorf("Pwd checker Check [%s] failed: password doesn't match regex", pc.field)
		return NewFailedResult("password doesn't match regex")
	}

	if err := passutils.CheckPassWordComplexity(&data); err != nil {
		hwlog.RunLog.Errorf("Pwd checker Check [%s] failed: the complex dose not meet the requirement", pc.field)
		return NewFailedResult("the complex dose not meet the requirement")
	}

	return NewSuccessResult()
}
