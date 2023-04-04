// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package checker

import (
	"fmt"

	"huawei.com/mindxedge/base/common/checker/valuer"
)

// GetExistChecker [method] for get exist checker
func GetExistChecker(field string) *ExistChecker {
	return &ExistChecker{
		field: field,
	}
}

// ExistChecker [struct] for exist checker
type ExistChecker struct {
	field  string
	valuer valuer.ExistValuer
}

// Check [method] for do exist check
func (bc *ExistChecker) Check(data interface{}) CheckResult {
	existFlag, err := bc.valuer.GetValue(data, bc.field)
	if existFlag {
		return NewSuccessResult()
	}
	return NewFailedResult(fmt.Sprintf("exist checker Check field [%s] failed, err:%v", bc.field, err))
}
