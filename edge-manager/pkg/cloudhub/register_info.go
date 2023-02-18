// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package cloudhub server init
package cloudhub

import (
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/websocketmgr"
)

var regInfoList = []websocketmgr.RegisterModuleInfo{
	{MsgOpt: common.OptGet, MsgRes: common.ResEdgeCoreConfig, ModuleName: common.NodeMsgManagerName},
	{MsgOpt: common.OptReport, MsgRes: common.ResDownloadProgressReport, ModuleName: common.NodeMsgManagerName},
	{MsgOpt: common.OptReport, MsgRes: common.ResSoftwareInfoReport, ModuleName: common.NodeManagerName},
	{MsgOpt: common.OptGet, MsgRes: common.ResDownLoadCert, ModuleName: common.NodeManagerName},
}

func getRegModuleInfoList() []websocketmgr.RegisterModuleInfo {
	return regInfoList
}
