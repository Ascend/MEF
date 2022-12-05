// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager database table
package appmanager

// AppInfo is app db table info
type AppInfo struct {
	ID          uint64 `gorm:"type:integer;primaryKey;autoIncrement:true"`
	AppName     string `gorm:"type:char(128);unique;not null"`
	Description string `gorm:"type:char(255);" json:"description"`
	CreatedAt   string `gorm:"type:char(19);not null"`
	ModifiedAt  string `gorm:"type:char(19);not null"`
	Containers  string `gorm:"type:text;not null" json:"containers"`
}

// AppInstance is application instance
type AppInstance struct {
	ID            uint64 `gorm:"type:Integer;primaryKey;autoIncrement:true"`
	PodName       string `gorm:"type:char(42);unique;not null"`
	NodeName      string `gorm:"type:char(64);not null"`
	NodeGroupName string `gorm:"type:char(64);not null"`
	NodeGroupID   int64  `gorm:"type:Integer;not null"`
	Status        string `gorm:"type:char(50);not null"`
	AppID         uint64 `gorm:"not null"`
	AppName       string `gorm:"not null"`
	CreatedAt     string `gorm:"type:time;not null"`
	ChangedAt     string `gorm:"type:time;not null"`
}
