// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package handlers this file for report module message handler register
package handlers

import (
	"sync"

	"huawei.com/mindx/common/modulemgr/handler"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-om/subalarm/handlers/alarm"
)

var handlerMgr handler.MsgHandler
var regOnce sync.Once
var registerInfoList = []handler.RegisterInfo{
	{MsgOpt: constants.OptReport, MsgRes: constants.ReportAlarmMsg, Handler: new(alarm.Handler)},
}

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
