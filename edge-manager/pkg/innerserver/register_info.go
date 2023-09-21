// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package innerserver

import (
	"huawei.com/mindx/common/websocketmgr"

	"huawei.com/mindxedge/base/common"
)

var regInfoList = []websocketmgr.RegisterModuleInfo{
	{MsgOpt: common.Get, MsgRes: common.GetSnsByGroup, ModuleName: common.NodeManagerName},
}

func getRegModuleInfoList() []websocketmgr.RegisterModuleInfo {
	return regInfoList
}
