// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package alarmmanager

import "time"

// AlarmStaticInfo is the struct for alarm_static_info table in the database
type AlarmStaticInfo struct {
	AlarmId             string `gorm:"type:varchar(64);primaryKey"`
	AlarmName           string `gorm:"type:varchar(64)"`
	PerceivedSeverity   string `gorm:"type:varchar(64);not null"`
	DetailedInformation string `gorm:"type:varchar(256)"`
	Suggestion          string `gorm:"type:varchar(4096)"`
	Reason              string `gorm:"type:varchar(256)"`
	Impact              string `gorm:"type:varchar(256)"`
	Resource            string `gorm:"type:varchar(256)"`
}

// AlarmInfo is the struct for alarm_info table in the database
type AlarmInfo struct {
	Id        int       `gorm:"primaryKey;autoIncrement"`
	AlarmType string    `gorm:"type:varchar(64);not null"`
	CreatedAt time.Time `gorm:"type:time;not null"`
	NodeId    int       `gorm:"type:int64;not null"`
	AlarmId   int       `gorm:"type:varchar(64);not null"`
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
	NodeId              int       `json:"nodeId"`
}
