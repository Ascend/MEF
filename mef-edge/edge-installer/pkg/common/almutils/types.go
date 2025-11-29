// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package almutils
package almutils

import (
	"context"
)

// alarm types
const (
	TypeAlarm = "alarm"
	TypeEvent = "event"
)

// notification types
const (
	NotifyTypeAlarm = "alarm"
	NotifyTypeClear = "clear"
	NotifyTypeEvent = ""
)

// severities
const (
	CRITICAL = "CRITICAL"
	MAJOR    = "MAJOR"
	MINOR    = "MINOR"
	OK       = "OK"
)

// AlarmMonitor alarm report interface
type AlarmMonitor interface {
	Monitoring(ctx context.Context)
	CollectOnce()
}

// Alarm defines an instance of alarm or event
type Alarm struct {
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

// Alarms defines a batch of Alarm
type Alarms struct {
	Alarm []Alarm `json:"alarm"`
}
