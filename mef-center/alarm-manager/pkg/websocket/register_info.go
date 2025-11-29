// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package websocket

import (
	"huawei.com/mindx/common/modulemgr"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
)

var regInfoList = []*modulemgr.RegisterModuleInfo{
	{MsgOpt: common.OptPost, MsgRes: requests.ReportAlarmRouter, ModuleName: common.AlarmManagerName},
	{MsgOpt: common.Delete, MsgRes: requests.ClearOneNodeAlarmRouter, ModuleName: common.AlarmManagerName},
}

func getRegModuleInfoList() []modulemgr.MessageHandlerIntf {
	handlers := make([]modulemgr.MessageHandlerIntf, len(regInfoList), len(regInfoList))
	for idx, reg := range regInfoList {
		handlers[idx] = reg
	}
	return handlers
}
