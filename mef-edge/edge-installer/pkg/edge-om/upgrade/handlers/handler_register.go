// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
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
