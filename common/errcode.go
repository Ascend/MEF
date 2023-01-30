// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package common to init error code
package common

const (
	// Success code
	Success = "00000000"
	// ErrorParseBody parse body failed
	ErrorParseBody = "00001001"
	// ErrorGetResponse get response failed
	ErrorGetResponse = "00001002"
	// ErrorsSendSyncMessageByRestful send sync message by restful failed
	ErrorsSendSyncMessageByRestful = "00001003"
	// ErrorResourceOptionNotFound module resource or option not found
	ErrorResourceOptionNotFound = "00001004"
	// ErrorParamInvalid parameter invalid
	ErrorParamInvalid = "00001005"
	// ErrorCreateAppTemplate failed to create app template
	ErrorCreateAppTemplate = "00002005"
	// ErrorDeleteAppTemplate failed to delete app template
	ErrorDeleteAppTemplate = "00002006"
	// ErrorModifyAppTemplate failed to modify app template
	ErrorModifyAppTemplate = "00002007"
	// ErrorGetAppTemplates failed to get app templates
	ErrorGetAppTemplates = "00002008"
	// ErrorGetAppTemplateDetail failed to get app template detail
	ErrorGetAppTemplateDetail = "00002009"

	// ErrorCreateUser fail to create user
	ErrorCreateUser = "10001001"
	// ErrorCreateUserToDb fail to insert user to db
	ErrorCreateUserToDb = "10001002"
	// ErrorChangePassword fail to modify password
	ErrorChangePassword = "10001003"
	// ErrorLogin fail to login
	ErrorLogin = "10001004"
	// ErrorNeedFirstLogin need first login
	ErrorNeedFirstLogin = "10001005"
	// ErrorLockState in lock state
	ErrorLockState = "10001006"
	// ErrorPasswordRepeat pass repeat in history
	ErrorPasswordRepeat = "10001008"
	// ErrorUserAlreadyFirstLogin user already first login
	ErrorUserAlreadyFirstLogin = "10001009"
	// ErrorPassOrUser username or password error
	ErrorPassOrUser = "10001010"
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
	// ErrorResourceOptionNotFound module resource or option not found info
	ErrorResourceOptionNotFound: "module resource or option not found",
	// ErrorParamInvalid parameter invalid info
	ErrorParamInvalid: "parameter invalid",
	// ErrorCreateAppTemplate failed to create app template info
	ErrorCreateAppTemplate: "failed to create app template",
	// ErrorDeleteAppTemplate failed to delete app template info
	ErrorDeleteAppTemplate: "failed to delete app template",
	// ErrorModifyAppTemplate failed to modify app template info
	ErrorModifyAppTemplate: "failed to modify app template",
	// ErrorGetAppTemplates failed to get app templates
	ErrorGetAppTemplates: "failed to get app templates",
	// ErrorGetAppTemplateDetail failed to get app template detail info
	ErrorGetAppTemplateDetail: "failed to get app template detail",

	// ErrorCreateUser fail to create user
	ErrorCreateUser: "failed to create user",
	// ErrorCreateUserToDb fail to insert user to db
	ErrorCreateUserToDb: "fail to insert user to db",
	// ErrorModifyPassword fail to modify password
	ErrorChangePassword: "fail to modify password",
	// ErrorLogin fail to login
	ErrorLogin: "fail to login",
	// ErrorNeedFirstLogin need first login
	ErrorNeedFirstLogin: "need first login",
	// ErrorLockState in lock state
	ErrorLockState: "user or ip in lock state",
	// ErrorPasswordRepeat password same with history
	ErrorPasswordRepeat: "password cannot be the same within recent 5 times",
	// ErrorUserAlreadyFirstLogin user already first login
	ErrorUserAlreadyFirstLogin: "user already first login",
	// ErrorPassOrUser username or password error
	ErrorPassOrUser: "username or password error",
}
