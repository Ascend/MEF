// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common base process used
package common

import (
	"encoding/json"
	"errors"
	"huawei.com/mindx/common/hwlog"
	"strings"
)

// ClearSliceByteMemory clears slice in memory
func ClearSliceByteMemory(sliceByte []byte) {
	for i := 0; i < len(sliceByte); i++ {
		sliceByte[i] = 0
	}
}

// ParamConvert convert request parameter from restful module
func ParamConvert(input interface{}, reqType interface{}) error {
	inputStr, ok := input.(string)
	if !ok {
		hwlog.RunLog.Error("create node conver request error1")
		return errors.New("convert requst error")
	}
	dec := json.NewDecoder(strings.NewReader(inputStr))
	if err := dec.Decode(reqType); err != nil {
		hwlog.RunLog.Error("create node conver request error3")
		return errors.New("decode requst error")
	}
	return nil
}
