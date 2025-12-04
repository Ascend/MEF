// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
