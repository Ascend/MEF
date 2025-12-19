// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package requests

const (
	// ReportAlarmRouter is the route that edgemanager forward the alarm report msg from MEFEdge to alarmmanager
	ReportAlarmRouter = "/edge/alarm/report"
	// ClearOneNodeAlarmRouter is the route that edgemanager send alarmmanager once one node is unmanaged
	ClearOneNodeAlarmRouter = "/edge/alarm/node-clear"
)

// AlarmsReq is the struct to deal alarms request
type AlarmsReq struct {
	Alarms []AlarmReq `json:"alarm"`
	Sn     string     `json:"serialNumber"`
	Ip     string     `json:"ip"`
}

// AlarmReq is the struct for one single alarm
type AlarmReq struct {
	Type                string `json:"type"`
	AlarmId             string `json:"alarmId"`
	AlarmName           string `json:"alarmName"`
	Resource            string `json:"resource"`
	PerceivedSeverity   string `json:"perceivedSeverity"`
	Timestamp           string `json:"timestamp"`
	NotificationType    string `json:"notificationType"`
	DetailedInformation string `json:"detailedInformation"`
	Suggestion          string `json:"suggestion"`
	Reason              string `json:"reason"`
	Impact              string `json:"impact"`
}

// ClearNodeAlarmReq is the struct to clear all alarm in one node
type ClearNodeAlarmReq struct {
	Sn string `json:"serialNumber"`
}

// GetSnsReq request for alarm-manager to get sns by nodegroup id
type GetSnsReq struct {
	GroupId uint64 `json:"groupId"`
}
