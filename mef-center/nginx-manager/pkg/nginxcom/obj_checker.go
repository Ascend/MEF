// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package nginxcom

import (
	"fmt"
	"reflect"
	"strconv"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/hwlog"
)

// ObjChecker [struct] for object checker
type ObjChecker struct {
	Checker  *checker.AndChecker
	DataType reflect.Kind
}

// Check [method] for do object check
func (oc *ObjChecker) Check(data string) checker.CheckResult {
	if oc.Checker == nil {
		return checker.NewFailedResult("object checker failed: the and checker not init")
	}
	dataItr, err := parseToData(oc.DataType, data)
	if err != nil {
		return checker.NewFailedResult("object checker failed: parse data failed")
	}
	ret := oc.Checker.Check(dataItr)
	if !ret.Result {
		hwlog.RunLog.Errorf("object checker failed: %s", ret.Reason)
	}
	return ret
}

func parseToData(dataType reflect.Kind, data string) (interface{}, error) {
	if dataType == reflect.Int {
		ret, err := strconv.Atoi(data)
		if err != nil {
			hwlog.RunLog.Error("convert data from string to int error")
			return nil, err
		}
		return ret, nil
	} else if dataType == reflect.String {
		return data, nil
	}
	hwlog.RunLog.Errorf("object checker not support type: %v", dataType)
	return nil, fmt.Errorf("object checker not support type: %v", dataType)
}
