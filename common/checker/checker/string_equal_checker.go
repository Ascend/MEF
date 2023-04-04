// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package checker for equal
package checker

import (
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common/checker/valuer"
)

// StringEqualChecker [struct] for string equal checker
type StringEqualChecker struct {
	field1 string
	field2 string
	valuer valuer.StringValuer
}

// GetStringEqualChecker [method] for get string equal checker
func GetStringEqualChecker(filed1 string, filed2 string) *StringEqualChecker {
	return &StringEqualChecker{
		field1: filed1,
		field2: filed2,
		valuer: valuer.StringValuer{},
	}
}

// Check [method] for do string equal check
func (ec *StringEqualChecker) Check(data interface{}) CheckResult {
	value1, err := ec.valuer.GetValue(data, ec.field1)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) {
			return NewFailedResult(fmt.Sprintf("field [%s] does not exist", ec.field1))
		}
		hwlog.RunLog.Errorf("String equal checker get field [%s] value failed, error: %v", ec.field1, err)
		return NewFailedResult(fmt.Sprintf("String equal checker get field [%s] value failed, error: %v",
			ec.field1, err))
	}

	value2, err := ec.valuer.GetValue(data, ec.field2)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) {
			return NewFailedResult(fmt.Sprintf("field [%s] does not exist", ec.field2))
		}
		hwlog.RunLog.Errorf("String equal checker get field [%s] value failed, error: %v", ec.field2, err)
		return NewFailedResult(fmt.Sprintf("String equal checker get field [%s] value failed, error: %v",
			ec.field2, err))
	}

	if value1 != value2 {
		hwlog.RunLog.Errorf("String equal checker Check failed: field [%s] value and field [%s] value are not equal",
			ec.field1, ec.field2)
		return NewFailedResult(fmt.Sprintf(
			"String equal checker Check failed: field [%s] value and field [%s] value are not equal",
			ec.field1, ec.field2))
	}

	return NewSuccessResult()
}
