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
	SerialNumber        string    `gorm:"type:varchar(64);not null;index:search_alarm"`
	AlarmId             string    `gorm:"type:varchar(64);not null;index:search_alarm"`
	AlarmName           string    `gorm:"type:varchar(64)"`
	PerceivedSeverity   string    `gorm:"type:varchar(64);not null"`
	DetailedInformation string    `gorm:"type:varchar(256)"`
	Suggestion          string    `gorm:"type:varchar(512)"`
	Reason              string    `gorm:"type:varchar(256)"`
	Impact              string    `gorm:"type:varchar(256)"`
	Resource            string    `gorm:"type:varchar(256)"`
}

func (ai *AlarmInfo) String() string {
	return fmt.Sprintf("AlarmInfo:  AlarmID:%d,AlarmType:%s,SerialNumber:%s,AlarmName:%s,PerceivedSeverity:%s,"+
		"Resource:%s", ai.Id, ai.AlarmType, ai.SerialNumber, ai.AlarmName, ai.PerceivedSeverity, ai.Resource)
}
