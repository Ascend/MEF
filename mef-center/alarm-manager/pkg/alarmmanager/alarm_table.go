// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package alarmmanager for alarm-manager module define tables
package alarmmanager

import (
	"time"
)

// AlarmInfo is the struct for alarm_info table in the database
type AlarmInfo struct {
	Id                  uint64    `gorm:"primaryKey;autoIncrement:true"                    json:"id"`
	AlarmType           string    `gorm:"type:varchar(64);not null"                        json:"alarmType"`
	CreatedAt           time.Time `gorm:"not null"                                         json:"createAt"`
	SerialNumber        string    `gorm:"type:varchar(64);not null;index:search_alarm"     json:"serialNumber"`
	Ip                  string    `gorm:"type:varchar(64);not null;"                       json:"ip"`
	AlarmId             string    `gorm:"type:varchar(64);not null;index:search_alarm"     json:"alarmId"`
	AlarmName           string    `gorm:"type:varchar(64)"                                 json:"alarmName"`
	PerceivedSeverity   string    `gorm:"type:varchar(64);not null"                        json:"perceivedSeverity"`
	DetailedInformation string    `gorm:"type:varchar(256)"                                json:"detailedInformation"`
	Suggestion          string    `gorm:"type:varchar(512)"                                json:"suggestion"`
	Reason              string    `gorm:"type:varchar(256)"                                json:"reason"`
	Impact              string    `gorm:"type:varchar(256)"                                json:"impact"`
	Resource            string    `gorm:"type:varchar(256)"                                json:"resource"`
}
