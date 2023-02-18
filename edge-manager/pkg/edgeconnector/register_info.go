// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeconnector defines register info
package edgeconnector

import (
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/websocketmgr"
)

var regInfoList = []websocketmgr.RegisterModuleInfo{
	{MsgOpt: common.OptGet, MsgRes: common.ResEdgeCoreConfig, ModuleName: common.EdgeInstallerName},
	{MsgOpt: common.OptGet, MsgRes: common.ResDownLoadSoftware, ModuleName: common.EdgeInstallerName},
	{MsgOpt: common.OptReport, MsgRes: common.ResProgressReport, ModuleName: common.EdgeInstallerName},
	{MsgOpt: common.OptGet, MsgRes: common.ResDownLoadCert, ModuleName: common.EdgeInstallerName},
}

func getRegModuleInfoList() []websocketmgr.RegisterModuleInfo {
	return regInfoList
}
