// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package checker for pwd
package checker

import (
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common/checker/valuer"
	"huawei.com/mindxedge/base/common/passutils"
)

// PwdChecker [struct] for password checker
type PwdChecker struct {
	fieldUser string
	fieldPwd  string
	maxLen    int
	minLen    int
	required  bool
	valuer    valuer.StringValuer
}

// GetPwdChecker [method] for get password checker
func GetPwdChecker(fieldUser, fieldPwd string, min, max int, required bool) *PwdChecker {
	return &PwdChecker{
		fieldUser: fieldUser,
		fieldPwd:  fieldPwd,
		minLen:    min,
		maxLen:    max,
		required:  required,
		valuer:    valuer.StringValuer{},
	}
}

// Check [method] for do password check
func (pc *PwdChecker) Check(data interface{}) CheckResult {
	name, err := pc.valuer.GetValue(data, pc.fieldUser)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) && !pc.required {
			return NewSuccessResult()
		}
		// todo 删除错误信息的返回
		hwlog.RunLog.Errorf("Pwd checker get field [%s] value failed, error: %v", pc.fieldUser, err)
		return NewFailedResult(fmt.Sprintf("get field [%s] value failed", pc.fieldUser))
	}

	value, err := pc.valuer.GetValue(data, pc.fieldPwd)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) && !pc.required {
			return NewSuccessResult()
		}
		// todo 删除错误信息的返回
		hwlog.RunLog.Errorf("Pwd checker get field [%s] value failed, error: %v", pc.fieldPwd, err)
		return NewFailedResult(fmt.Sprintf("get field [%s] value failed", pc.fieldPwd))
	}

	return pc.isPwdValid(name, value)
}

func (pc *PwdChecker) isPwdValid(name, value string) CheckResult {
	if len(value) < pc.minLen || len(value) > pc.maxLen {
		hwlog.RunLog.Errorf("Pwd checker Check [%s] failed: the length is not in range [%d, %d]",
			pc.fieldPwd, pc.minLen, pc.maxLen)
		return NewFailedResult(fmt.Sprintf("the length is not in range [%d, %d]", pc.minLen, pc.maxLen))
	}

	if err := passutils.CheckPassWord(name, &value); err != nil {
		hwlog.RunLog.Errorf("Pwd checker Check [%s] failed, error: %v", pc.fieldPwd, err)
		return NewFailedResult(err.Error())
	}

	return NewSuccessResult()
}
