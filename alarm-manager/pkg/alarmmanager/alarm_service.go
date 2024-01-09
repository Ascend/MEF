// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package alarmmanager for alarm-manager module handler
package alarmmanager

import (
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindxedge/base/common/alarms"
)

func listAlarms(msg *model.Message) (interface{}, error) {
	return dealRequest(msg, alarms.AlarmType), nil
}

func listEvents(msg *model.Message) (interface{}, error) {
	return dealRequest(msg, alarms.EventType), nil
}

func getAlarmDetail(msg *model.Message) (interface{}, error) {
	return getAlarmOrEventDbDetail(msg, alarms.AlarmType), nil
}

func getEventDetail(msg *model.Message) (interface{}, error) {
	return getAlarmOrEventDbDetail(msg, alarms.EventType), nil
}
