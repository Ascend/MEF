// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package cloudhub server init
package cloudhub

import (
	"huawei.com/mindx/common/websocketmgr"

	"edge-manager/pkg/constants"
	"huawei.com/mindxedge/base/common"
)

var regInfoList = []websocketmgr.RegisterModuleInfo{
	{MsgOpt: common.OptGet, MsgRes: common.ResConfig, ModuleName: common.NodeMsgManagerName},
	{MsgOpt: common.OptReport, MsgRes: common.ResDownloadProgress, ModuleName: common.NodeMsgManagerName},
	{MsgOpt: common.OptReport, MsgRes: common.ResSoftwareInfo, ModuleName: common.NodeManagerName},
	{MsgOpt: common.OptGet, MsgRes: common.ResDownLoadCert, ModuleName: common.NodeMsgManagerName},
	{MsgOpt: common.OptPost, MsgRes: common.ResEdgeCert, ModuleName: common.CloudHubName},
	{MsgOpt: common.OptGet, MsgRes: common.CertWillOverdue, ModuleName: common.CloudHubName},
	{MsgOpt: common.Delete, MsgRes: common.DeleteNodeMsg, ModuleName: common.NodeManagerName},
	{MsgOpt: common.OptReport, MsgRes: constants.ResLogDumpError, ModuleName: constants.LogManagerName},
}

func getRegModuleInfoList() []websocketmgr.RegisterModuleInfo {
	return regInfoList
}
