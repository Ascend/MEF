// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package handlers this file for handlers register
package handlers

import (
	"sync"

	"huawei.com/mindx/common/modulemgr/handler"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-om/upgrade/handlers/upgrade"
	"edge-installer/pkg/edge-om/upgrade/handlers/verification"
	"edge-installer/pkg/edge-om/upgrade/reporter"
)

var handlerMgr handler.MsgHandler
var regOnce sync.Once
var registerInfoList = []handler.RegisterInfo{
	{MsgOpt: constants.OptPost, MsgRes: constants.InnerSoftwareVerification, Handler: new(verification.Handler)},
	{MsgOpt: constants.OptPost, MsgRes: constants.ResUpgradeInfo, Handler: new(upgrade.Handler)},
	{MsgOpt: constants.OptReport, MsgRes: constants.InnerSoftwareVersion, Handler: new(reporter.Handler)},
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
