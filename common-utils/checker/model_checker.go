// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package checker

import (
	"fmt"

	"huawei.com/mindx/common/checker/valuer"
)

// ModelChecker [struct] for model checkers
type ModelChecker struct {
	Field    string
	Required bool
	Checker  checkerIntf
}

// Check [method] for do model check
func (mc *ModelChecker) Check(data interface{}) CheckResult {
	if mc.Checker == nil {
		return NewFailedResult("model checker failed: the and checker not init")
	}

	if mc.Field == "" {
		return mc.Checker.Check(data)
	}

	value, err := valuer.GetReflectValueByName(data, mc.Field)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) && !mc.Required {
			return NewSuccessResult()
		}
		return NewFailedResult(fmt.Sprintf("field [%s] not find", mc.Field))
	}
	return mc.Checker.Check(value)
}
