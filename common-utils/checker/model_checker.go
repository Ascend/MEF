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
