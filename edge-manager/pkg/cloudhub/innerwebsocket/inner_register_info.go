// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package innerwebsocket

import (
	"fmt"

	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"
)

var regInfoList = []*modulemgr.RegisterModuleInfo{
	{MsgOpt: common.Get, MsgRes: common.GetSnsByGroup, ModuleName: common.NodeManagerName},
}

func getRegModuleInfoList() []modulemgr.MessageHandlerIntf {
	handlers := make([]modulemgr.MessageHandlerIntf, len(regInfoList), len(regInfoList))
	for idx, reg := range regInfoList {
		handlers[idx] = reg
	}
	return handlers
}

// AlarmClearHandler handler for requesting alarm manager to clear an alarm though the inner ws link
func AlarmClearHandler(message *model.Message) (*model.Message, bool, error) {
	if err := sendMessageByInnerWs(message, common.AlarmManagerWsMoudle); err != nil {
		return message, false, fmt.Errorf("send ws msg to %s failed: %s", message.GetNodeId(), err.Error())
	}
	return message, false, nil
}

// AlarmReportHandler handler for edge node or center reporting alarms to alarm-manager pod though the inner ws link
func AlarmReportHandler(message *model.Message) (*model.Message, bool, error) {
	modifyMsgForAlarmManager(message)
	if err := sendMessageByInnerWs(message, common.AlarmManagerWsMoudle); err != nil {
		// edge-node will retry by interval, so if error happens will not deal with it
		return message, false, fmt.Errorf("send ws msg to %s failed: %s", message.GetNodeId(), err.Error())
	}
	return message, false, nil
}
