// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package subalarm this file for report module message handler register
package subalarm

import (
	"sync"

	"huawei.com/mindx/common/modulemgr/handler"
)

var handlerMgr handler.MsgHandler
var regOnce sync.Once
var registerInfoList []handler.RegisterInfo

// GetHandlerMgr get handler manager
func GetHandlerMgr() *handler.MsgHandler {
	regOnce.Do(registerHandler)
	return &handlerMgr
}

func registerHandler() {
	handlerMgr = handler.MsgHandler{}
	for _, reg := range registerInfoList {
		handlerMgr.Register(reg)
	}
}
