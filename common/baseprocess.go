// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common base process used
package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unsafe"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/rand"
)

// Router struct
type Router struct {
	Source      string
	Destination string
	Option      string
	Resource    string
}

const (
	sensitiveInfoWildcard = "***"
	// minimum substring length to be replaced
	minimumCommonSubStrLen = 2
	maxRandomLen           = 10240
	minSaltLen             = 16
)

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
func SendSyncMessageByRestful(input interface{}, router *Router, timeout time.Duration) RespMsg {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Error("new message error")
		return RespMsg{Status: ErrorsSendSyncMessageByRestful, Msg: "", Data: nil}
	}

	msg.SetRouter(router.Source, router.Destination, router.Option, router.Resource)
	msg.FillContent(input)

	respMsg, err := modulemgr.SendSyncMessage(msg, timeout)
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

// TrimInfoFromError trim sensitive information from an error, return new error
func TrimInfoFromError(err error, sensitiveStr string) error {
	if err == nil || sensitiveStr == "" {
		return err
	}
	oldErrStr := err.Error()
	if oldErrStr == "" {
		return err
	}
	if strings.Contains(sensitiveStr, oldErrStr) {
		return fmt.Errorf(sensitiveInfoWildcard)
	}
	newErrStr := strings.ReplaceAll(oldErrStr, sensitiveStr, sensitiveInfoWildcard)
	if newErrStr != oldErrStr {
		return fmt.Errorf(newErrStr)
	}
	commonStr := MaxCommonSubStr(sensitiveStr, oldErrStr)
	if len(commonStr) >= minimumCommonSubStrLen {
		newErrStr = strings.ReplaceAll(oldErrStr, commonStr, sensitiveInfoWildcard)
	}
	return fmt.Errorf(newErrStr)
}

// MaxCommonSubStr get the max common substring between two strings.
func MaxCommonSubStr(s1 string, s2 string) string {
	if len(s1) == 0 || len(s2) == 0 {
		return ""
	}
	str1Len := len(s1)
	str2Len := len(s2)
	a := make([][]int, str1Len)
	for i := 0; i < len(a); i++ {
		a[i] = make([]int, str2Len)
		if []byte(s1)[i] == []byte(s2)[0] {
			a[i][0] = 1
		}
	}
	for j := 1; j < str2Len; j++ {
		if []byte(s1)[0] == []byte(s2)[j] {
			a[0][j] = 1
		}
	}
	max := 0
	idx1 := 0
	idx2 := 0
	for i := 1; i < len(a); i++ {
		for j := 1; j < str2Len; j++ {
			if []byte(s1)[i] == []byte(s2)[j] {
				a[i][j] = a[i-1][j-1] + 1
			}
			if a[i][j] > max {
				max = a[i][j]
				idx1 = i
				idx2 = j
			}
		}
	}
	return string([]byte(s1)[idx1+1-a[idx1][idx2] : idx1+1])
}

// GetSafeRandomBytes get safe random bytes
func GetSafeRandomBytes(saltLength int) ([]byte, error) {
	if saltLength > maxRandomLen || saltLength < minSaltLen {
		return nil, errors.New("salt length invalid")
	}
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}
