// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package innerclient this file for register module route transfer
package innerclient

import (
	"huawei.com/mindx/common/modulemgr"

	"edge-installer/pkg/common/constants"
)

var registerInfoList = []*modulemgr.RegisterModuleInfo{
	{MsgOpt: constants.OptGet, MsgRes: constants.ResConfig, ModuleName: constants.ModEdgeOm},
	{MsgOpt: constants.OptUpdate, MsgRes: constants.ResImageCertInfo, ModuleName: constants.ModEdgeOm},
	{MsgOpt: constants.OptUpdate, MsgRes: constants.ResNpuSharing, ModuleName: constants.ModEdgeOm},
	{MsgOpt: constants.OptRestart, MsgRes: constants.ActionPod, ModuleName: constants.ModEdgeOm},
	{MsgOpt: constants.OptReport, MsgRes: constants.DeviceOmConnectMsg, ModuleName: constants.OmJobManager},
	{MsgOpt: constants.OptReport, MsgRes: constants.ReportAlarmMsg, ModuleName: constants.OmAlarmMgr},
	{MsgOpt: constants.OptUpdate, MsgRes: constants.ActionModelFiles, ModuleName: constants.ModEdgeOm},
	{MsgOpt: constants.OptRaw, MsgRes: constants.ActionModelFiles, ModuleName: constants.ModEdgeOm},
}

func getRegModuleInfoList() []modulemgr.MessageHandlerIntf {
	handlers := make([]modulemgr.MessageHandlerIntf, 0, len(registerInfoList))
	for _, reg := range registerInfoList {
		handlers = append(handlers, reg)
	}
	return handlers
}
