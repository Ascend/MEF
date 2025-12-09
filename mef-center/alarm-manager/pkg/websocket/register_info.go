// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
