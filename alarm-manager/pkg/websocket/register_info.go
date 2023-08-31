// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package websocket

import (
	"alarm-manager/pkg/alarmmanager"

	"huawei.com/mindx/common/websocketmgr"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
)

var regInfoList = []websocketmgr.RegisterModuleInfo{
	{MsgOpt: common.OptPost, MsgRes: requests.ReportAlarmRouter, ModuleName: alarmmanager.AlarmModuleName},
	{MsgOpt: common.Delete, MsgRes: requests.ClearOneNodeAlarmRouter, ModuleName: alarmmanager.AlarmModuleName},
}

func getRegModuleInfoList() []websocketmgr.RegisterModuleInfo {
	return regInfoList
}
