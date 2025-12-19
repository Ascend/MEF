// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package logmgmt provides utils for logging
package logmgmt

import (
	"encoding/json"
	"fmt"
	"reflect"

	"huawei.com/mindx/common/hwlog"
)

// BatchOperationLog is the func to record the success log for batch operation
func BatchOperationLog(prefix string, retList []interface{}) {
	var logContent string
	const (
		maxLength     = 1024
		separator     = ", "
		separationLen = 2
	)
	logStart := fmt.Sprintf("%s [", prefix)
	logEnd := "] success"
	for _, ret := range retList {
		content := parseContent(ret) + separator
		totalLength := len(logStart) + len(logEnd) + len(logContent) + len(content) - separationLen
		if totalLength > maxLength && len(logContent) > separationLen {
			logContent = logContent[:len(logContent)-separationLen]
			hwlog.RunLog.Infof("%s%s%s", logStart, logContent, logEnd)
			logContent = ""
		}
		logContent += content
	}
	if len(logContent) < separationLen {
		return
	}
	logContent = logContent[:len(logContent)-separationLen]
	hwlog.RunLog.Infof("%s%s%s", logStart, logContent, logEnd)
}

func parseContent(content interface{}) string {
	contentType := reflect.TypeOf(content)
	if contentType == nil {
		return ""
	}
	switch contentType.Kind() {
	case reflect.Struct:
		ret, err := json.Marshal(content)
		if err != nil {
			return ""
		}
		return string(ret)
	case reflect.Ptr:
		return parseContent(reflect.ValueOf(content).Elem().Interface())
	default:
		return fmt.Sprintf("%v", content)
	}
}
