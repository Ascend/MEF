// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common base process used
package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

// Router struct
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

// ClearStringMemory clears string in memory
func ClearStringMemory(s string) {
	bs := *(*[]byte)(unsafe.Pointer(&s))
	for i := 0; i < len(bs); i++ {
		bs[i] = 0
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
		hwlog.RunLog.Error("param type is not string")
		return errors.New("param type error")
	}
	dec := json.NewDecoder(strings.NewReader(inputStr))
	if err := dec.Decode(reqType); err != nil {
		hwlog.RunLog.Errorf("param decode failed: %s", err.Error())
		return errors.New("param decode error")
	}
	return nil
}

// Combine to combine option and resource to find url method
func Combine(option, resource string) string {
	return fmt.Sprintf("%s%s", option, resource)
}
