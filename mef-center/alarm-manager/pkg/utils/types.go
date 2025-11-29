// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package utils for alarm manager
package utils

import (
	"time"
)

// ListAlarmOrEventReq can not have both GroupId and Sn
type ListAlarmOrEventReq struct {
	PageNum  uint64 `json:"pageNum"`
	PageSize uint64 `json:"pageSize"`
	Sn       string `json:"serialNumber,omitempty"`
	GroupId  uint64 `json:"groupId,omitempty"`
	IfCenter string `json:"ifCenter,omitempty"`
}

// AlarmBriefInfo the simple information for respond to User
type AlarmBriefInfo struct {
	ID        uint64    `json:"id"`
	Sn        string    `json:"serialNumber"`
	Ip        string    `json:"ip"`
	Severity  string    `json:"severity"`
	Resource  string    `json:"resource"`
	CreatedAt time.Time `json:"createAt"`
	AlarmType string    `json:"alarmType"`
}

// ListAlarmsResp return list of resp for list alarms
type ListAlarmsResp struct {
	// Records alarm or events records info
	Records []AlarmBriefInfo `json:"records"`
	// Total is num of alarmInfos
	Total int64 `json:"total"`
}
