// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package alarmmanager for alarm-manager module define tables
package alarmmanager

import (
	"time"
)

// AlarmInfo is the struct for alarm_info table in the database
type AlarmInfo struct {
	Id                  uint64    `gorm:"primaryKey;autoIncrement:true"                    json:"id"`
	AlarmType           string    `gorm:"type:varchar(64);not null"                        json:"alarmType"`
	CreatedAt           time.Time `gorm:"type:time;not null"                               json:"createAt"`
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
