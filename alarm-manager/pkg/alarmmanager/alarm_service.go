// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package alarmmanager for alarm-manager module handler
package alarmmanager

import "alarm-manager/pkg/utils"

const (
	centerNodeQueryType    = "CenterNodeQuery"
	nodeIdQueryType        = "NodeIdQuery"
	groupIdQueryType       = "GroupIdQuery"
	fullNodesQueryType     = "FullNodesQuery"
	fullEdgeNodesQueryType = "FullEdgeNodesQuery"
)

func listAlarms(input interface{}) (interface{}, error) {
	return dealRequest(input, utils.AlarmType), nil
}

func listEvents(input interface{}) (interface{}, error) {
	return dealRequest(input, utils.EventType), nil
}

func getAlarmDetail(input interface{}) (interface{}, error) {
	return getAlarmOrEventDbDetail(input, utils.AlarmType), nil
}

func getEventDetail(input interface{}) (interface{}, error) {
	return getAlarmOrEventDbDetail(input, utils.EventType), nil
}
