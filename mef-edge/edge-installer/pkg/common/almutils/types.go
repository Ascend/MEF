// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
