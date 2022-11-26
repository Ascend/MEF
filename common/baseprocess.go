// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common base process used
package common

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

// Router router struct
type Router struct {
	Source      string
	Destination string
	Option      string
	Resource    string
}

// ClearSliceByteMemory clears slice in memory
func ClearSliceByteMemory(sliceByte []byte) {
	for i := 0; i < len(sliceByte); i++ {
		sliceByte[i] = 0
	}
}

// SendSyncMessageByRestful send sync message by restful
func SendSyncMessageByRestful(input interface{}, router *Router) RespMsg {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Error("new message error")
		return RespMsg{Status: ErrorsSendSyncMessageByRestful, Msg: "", Data: nil}
	}
	msg.SetRouter(router.Source, router.Destination, router.Option, router.Resource)
	msg.FillContent(input)
	respMsg, err := modulemanager.SendSyncMessage(msg, ResponseTimeout)
	if err != nil {
		hwlog.RunLog.Error("get response error")
		return RespMsg{Status: ErrorsSendSyncMessageByRestful, Msg: "", Data: nil}
	}
	return marshalResponse(respMsg)
}

func marshalResponse(respMsg *model.Message) RespMsg {
	content := respMsg.GetContent()
	respStr, err := json.Marshal(content)
	if err != nil {
		return RespMsg{Status: ErrorGetResponse, Msg: "", Data: nil}
	}
	var resp RespMsg
	if err := json.Unmarshal(respStr, &resp); err != nil {
		return RespMsg{Status: ErrorGetResponse, Msg: "", Data: nil}
	}
	return resp
}

// ParamConvert convert request parameter from restful module
func ParamConvert(input interface{}, reqType interface{}) error {
	inputStr, ok := input.(string)
	if !ok {
		hwlog.RunLog.Error("create node convert request error1")
		return errors.New("convert request error")
	}
	dec := json.NewDecoder(strings.NewReader(inputStr))
	if err := dec.Decode(reqType); err != nil {
		hwlog.RunLog.Error("create node convert request error3")
		return errors.New("decode request error")
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

// GetEdgeMgrWorkPath gets edge-manager work path
func GetEdgeMgrWorkPath() (string, bool) {
	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		hwlog.RunLog.Errorf("get edge-manager work absolute path error: %v", err)
		return "", false
	}

	currentDir, err = filepath.EvalSymlinks(currentDir)
	if err != nil {
		hwlog.RunLog.Errorf("get edge-manager work real path error: %v", err)
		return "", false
	}

	return currentDir, true
}
