// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.
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
