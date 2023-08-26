// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package alarmmanager for alarm-manager module define tables
package alarmmanager

import (
	"fmt"
	"time"
)

// AlarmInfo is the struct for alarm_info table in the database
type AlarmInfo struct {
	Id                  uint64    `gorm:"primaryKey;autoIncrement:true"`
	AlarmType           string    `gorm:"type:varchar(64);not null"`
	CreatedAt           time.Time `gorm:"type:time;not null"`
	NodeId              uint64    `gorm:"type:int64;not null"`
	AlarmId             string    `gorm:"type:varchar(64);not null"`
	AlarmName           string    `gorm:"type:varchar(64)"`
	PerceivedSeverity   string    `gorm:"type:varchar(64);not null"`
	DetailedInformation string    `gorm:"type:varchar(256)"`
	Suggestion          string    `gorm:"type:varchar(256)"`
	Reason              string    `gorm:"type:varchar(256)"`
	Impact              string    `gorm:"type:varchar(256)"`
	Resource            string    `gorm:"type:varchar(256)"`
}

func (ai *AlarmInfo) String() string {
	return fmt.Sprintf("AlarmInfo:  AlarmID:%d,AlarmType:%s,NodeId:%d,AlarmName:%s,PerceivedSeverity:%s,"+
		"Resource:%s", ai.Id, ai.AlarmType, ai.NodeId, ai.AlarmName, ai.PerceivedSeverity, ai.Resource)
}

type addAlarmReq struct {
	Alarms []oneAlarmReq `json:"alarms"`
}

type oneAlarmReq struct {
	AlarmId             string    `json:"alarmId"`
	AlarmName           string    `json:"alarmName"`
	Resource            string    `json:"resource"`
	NotificationType    string    `json:"notificationType"`
	PerceivedSeverity   string    `json:"perceivedSeverity"`
	DetailedInformation string    `json:"detailedInformation"`
	Suggestion          string    `json:"suggestion"`
	Reason              string    `json:"reason"`
	Impact              string    `json:"impact"`
	Type                string    `json:"type"`
	CreatedAt           time.Time `json:"timestamp"`
	NodeId              uint64    `json:"nodeId"`
}
