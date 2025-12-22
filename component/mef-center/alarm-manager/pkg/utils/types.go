// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
