// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common base process used
package common

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
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

// BindUriWithJSON convert uri to key-value string dict
func BindUriWithJSON(c *gin.Context) ([]byte, error) {
	if c == nil {
		return nil, errors.New("gin Context can't be nil")
	}
	obj := make(map[string][]string, len(c.Params))
	for _, v := range c.Params {
		params, ok := obj[v.Key]
		if !ok {
			params = make([]string, 0, 1)
		}
		obj[v.Key] = append(params, v.Value)
	}
	return json.Marshal(obj)
}
