// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package innerserver

import (
	"huawei.com/mindx/common/websocketmgr"
)

var regInfoList = []websocketmgr.RegisterModuleInfo{}

func getRegModuleInfoList() []websocketmgr.RegisterModuleInfo {
	return regInfoList
}
