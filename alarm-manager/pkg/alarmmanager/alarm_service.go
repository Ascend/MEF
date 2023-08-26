// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package alarmmanager for alarm-manager module handler
package alarmmanager

import (
	"huawei.com/mindxedge/base/common"
)

const (
	centerNodeQueryType    = "CenterNodeQuery"
	nodeIdQueryType        = "NodeIdQuery"
	groupIdQueryType       = "GroupIdQuery"
	fullNodesQueryType     = "FullNodesQuery"
	fullEdgeNodesQueryType = "FullEdgeNodesQuery"
)

func listAlarms(input interface{}) common.RespMsg {
	return dealRequest(input, AlarmType)
}

func listEvents(input interface{}) common.RespMsg {
	return dealRequest(input, EventType)
}

func getAlarmDetail(input interface{}) common.RespMsg {
	return getAlarmOrEventDbDetail(input, AlarmType)
}

func getEventDetail(input interface{}) common.RespMsg {
	return getAlarmOrEventDbDetail(input, EventType)
}
