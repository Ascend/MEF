// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
