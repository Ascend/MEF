// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package cloudhub server init
package cloudhub

import (
	"huawei.com/mindx/common/websocketmgr"

	"huawei.com/mindxedge/base/common"
)

var regInfoList = []websocketmgr.RegisterModuleInfo{
	{MsgOpt: common.OptGet, MsgRes: common.ResEdgeCoreConfig, ModuleName: common.NodeMsgManagerName},
	{MsgOpt: common.OptGet, MsgRes: common.ResConfig, ModuleName: common.NodeMsgManagerName},
	{MsgOpt: common.OptReport, MsgRes: common.ResDownloadProgress, ModuleName: common.NodeMsgManagerName},
	{MsgOpt: common.OptReport, MsgRes: common.ResSoftwareInfo, ModuleName: common.NodeManagerName},
	{MsgOpt: common.OptGet, MsgRes: common.ResDownLoadCert, ModuleName: common.NodeMsgManagerName},
}

func getRegModuleInfoList() []websocketmgr.RegisterModuleInfo {
	return regInfoList
}
