// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package requests

const (
	// ReportAlarmRouter is the route that edgemanager forward the alarm report msg from MEFEdge to alarmmanager
	ReportAlarmRouter = "/edge/alarm/report"
	// ClearOneNodeAlarmRouter is the route that edgemanager send alarmmanager once one node is unmanaged
	ClearOneNodeAlarmRouter = "/edge/alarm/node-clear"
)

// AddAlarmReq is the struct to add an alarm into alarm-manager
type AddAlarmReq struct {
	Alarms []AlarmReq `json:"alarm"`
	Sn     string     `json:"serialNumber"`
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
