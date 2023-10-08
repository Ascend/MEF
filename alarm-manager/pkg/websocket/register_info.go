// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package websocket

import (
	"huawei.com/mindx/common/websocketmgr"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
)

var regInfoList = []websocketmgr.RegisterModuleInfo{
	{MsgOpt: common.OptPost, MsgRes: requests.ReportAlarmRouter, ModuleName: common.AlarmManagerName},
	{MsgOpt: common.Delete, MsgRes: requests.ClearOneNodeAlarmRouter, ModuleName: common.AlarmManagerName},
}

func getRegModuleInfoList() []websocketmgr.RegisterModuleInfo {
	return regInfoList
}
