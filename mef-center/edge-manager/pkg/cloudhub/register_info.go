// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package cloudhub server init
package cloudhub

import (
	"huawei.com/mindx/common/modulemgr"

	"edge-manager/pkg/constants"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
)

const (
	alarmHandlerRate     = 35
	alarmHandlerCapacity = 90

	defaultHandlerRate     = 16
	defaultHandlerCapacity = 2048
)

var regInfoList = []*modulemgr.RegisterModuleInfo{
	{MsgOpt: common.OptGet, MsgRes: common.ResConfig, ModuleName: common.NodeMsgManagerName},
	{MsgOpt: common.OptReport, MsgRes: common.ResDownloadProgress, ModuleName: common.NodeMsgManagerName},
	{MsgOpt: common.OptReport, MsgRes: common.ResSoftwareInfo, ModuleName: common.NodeManagerName},
	{MsgOpt: common.OptGet, MsgRes: common.ResDownLoadCert, ModuleName: common.NodeMsgManagerName},
	{MsgOpt: common.OptPost, MsgRes: common.ResEdgeCert, ModuleName: common.CloudHubName},
	{MsgOpt: common.OptResp, MsgRes: common.CertWillExpired, ModuleName: common.CertUpdaterName},
	{MsgOpt: common.OptReport, MsgRes: constants.ResLogDumpError, ModuleName: constants.LogManagerName},
	{MsgOpt: common.OptPost, MsgRes: requests.ReportAlarmRouter, ModuleName: common.CloudHubName,
		Rps: alarmHandlerRate, Burst: alarmHandlerCapacity},
}

func getRegModuleInfoList() []modulemgr.MessageHandlerIntf {
	handlers := make([]modulemgr.MessageHandlerIntf, len(regInfoList), len(regInfoList))
	for idx, reg := range regInfoList {
		if reg.Rps == 0 {
			reg.Rps, reg.Burst = defaultHandlerRate, defaultHandlerCapacity
		}
		handlers[idx] = reg
	}
	return handlers
}
