// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package types for
package types

import (
	"fmt"
	"time"
)

// ListAlarmOrEventReq can not have both GroupId and NodeId
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

func (dig AlarmBriefInfo) String() string {
	return fmt.Sprintf("DigestInfo  ID: %d,SerialNumber:%s,Severity:%s,Resource:%s,CreateTime:%v", dig.ID,
		dig.Sn, dig.Severity, dig.Resource, dig.CreatedAt)
}
