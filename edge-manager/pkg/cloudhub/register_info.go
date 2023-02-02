// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package cloudhub defines register info
package cloudhub

import (
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/websocketmgr"
)

var regInfoList = []websocketmgr.RegisterModuleInfo{
	{MsgOpt: common.OptGet, MsgRes: common.ResEdgeCoreConfig, ModuleName: common.NodeMsgManagerName},
	{MsgOpt: common.OptPost, MsgRes: common.ResProgressReport, ModuleName: common.NodeMsgManagerName},
	{MsgOpt: common.OptPost, MsgRes: common.ResSoftwareInfoReport, ModuleName: common.NodeMsgManagerName},
}

func getRegModuleInfoList() []websocketmgr.RegisterModuleInfo {
	return regInfoList
}
