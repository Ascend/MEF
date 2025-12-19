// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package handlers
package handlers

import (
	"sync"

	"huawei.com/mindx/common/modulemgr/handler"

	"edge-installer/pkg/common/constants"
)

var msgHandler handler.MsgHandler
var regOnce sync.Once
var registerInfoList = []handler.RegisterInfo{
	{MsgOpt: constants.OptRestart, MsgRes: constants.ActionPod, Handler: new(podRestartHandler)},
	{MsgOpt: constants.OptUpdate, MsgRes: constants.ActionModelFiles, Handler: new(updateModelFileHandler)},
	{MsgOpt: constants.OptUpdate, MsgRes: constants.ActionContainerInfo, Handler: new(updateContainerInfoHandler)},
	{MsgOpt: constants.OptDelete, MsgRes: constants.ActionPodsData, Handler: new(deleteModelFileHandler)},
	{MsgOpt: constants.OptInsert, MsgRes: constants.ActionDefaultNodeStatus, Handler: new(nodeResourceEventHandler)},
	{MsgOpt: constants.OptPatch, MsgRes: constants.ActionDefaultNodePatch, Handler: new(nodeResourceEventHandler)},
	{MsgOpt: constants.OptPatch, MsgRes: constants.ActionPodPatch, Handler: new(podRestartEventHandler)},
}

// GetHandler get message handler
func GetHandler() *handler.MsgHandler {
	regOnce.Do(registerHandler)
	return &msgHandler
}

func registerHandler() {
	for _, reg := range registerInfoList {
		msgHandler.Register(reg)
	}
}
