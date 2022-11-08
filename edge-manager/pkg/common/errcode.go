// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package common to init error code
package common

const (
	// Success success code
	Success = "00000000"
	// ErrorParseBody parse body failed
	ErrorParseBody = "00001001"
	// ErrorGetResponse get response failed
	ErrorGetResponse = "00001002"
	// ErrorsSendSyncMessageByRestful send sync message by restful failed
	ErrorsSendSyncMessageByRestful = "00001003"
)

// ErrorMap error code and error msg map
var ErrorMap = map[string]string{
	// Success success code
	Success: "success",
	// ErrorParseBody parse body failed
	ErrorParseBody: "parse request body failed",
	// ErrorGetResponse get response failed
	ErrorGetResponse: "get response failed",
	// ErrorsSendSyncMessageByRestful send sync message by restful failed
	ErrorsSendSyncMessageByRestful: "send sync message by restful failed",
}
