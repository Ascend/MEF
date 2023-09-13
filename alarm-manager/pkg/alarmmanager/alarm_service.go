// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package alarmmanager for alarm-manager module handler
package alarmmanager

import (
	"huawei.com/mindxedge/base/common/alarms"
)

const (
	centerNodeQueryType    = "CenterNodeQuery"
	serialNumQuery         = "SerialNumQuery"
	groupIdQueryType       = "GroupIdQuery"
	fullNodesQueryType     = "FullNodesQuery"
	fullEdgeNodesQueryType = "FullEdgeNodesQuery"
)

func listAlarms(input interface{}) (interface{}, error) {
	return dealRequest(input, alarms.AlarmType), nil
}

func listEvents(input interface{}) (interface{}, error) {
	return dealRequest(input, alarms.EventType), nil
}

func getAlarmDetail(input interface{}) (interface{}, error) {
	return getAlarmOrEventDbDetail(input, alarms.AlarmType), nil
}

func getEventDetail(input interface{}) (interface{}, error) {
	return getAlarmOrEventDbDetail(input, alarms.EventType), nil
}
