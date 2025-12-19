// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package edgehub this file for register module routing
package edgehub

import (
	"huawei.com/mindx/common/modulemgr"

	"edge-installer/pkg/common/constants"
)

var regInfoList = []*modulemgr.RegisterModuleInfo{
	{MsgOpt: constants.OptPost, MsgRes: constants.ResDownloadCert, ModuleName: constants.ModEdgeOm},
	{MsgOpt: constants.OptPost, MsgRes: constants.ResEdgeDownloadInfo, ModuleName: constants.DownloadManagerName},
	{MsgOpt: constants.OptPost, MsgRes: constants.ResUpgradeInfo, ModuleName: constants.ModEdgeOm},
	{MsgOpt: constants.OptGet, MsgRes: constants.ResCertUpdate, ModuleName: constants.ModEdgeHub},
	{MsgOpt: constants.OptDelete, MsgRes: constants.DeleteNodeMsg, ModuleName: constants.ModEdgeHub},
	{MsgOpt: constants.OptPost, MsgRes: constants.ResDumpLogTask, ModuleName: constants.ModHandlerMgr},
}

func getRegModuleInfoList() []modulemgr.MessageHandlerIntf {
	handlers := make([]modulemgr.MessageHandlerIntf, 0, len(regInfoList))
	for _, reg := range regInfoList {
		handlers = append(handlers, reg)
	}
	return handlers
}
