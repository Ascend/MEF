// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager database table
package appmanager

import "time"

// AppInfo is app db table info
type AppInfo struct {
	ID          uint64    `gorm:"type:integer;primaryKey;autoIncrement:true"`
	AppName     string    `gorm:"type:char(128);unique;not null"`
	Description string    `gorm:"type:char(255);" json:"description"`
	CreatedAt   time.Time `gorm:"type:time"`
	UpdatedAt   time.Time `gorm:"type:time"`
	Containers  string    `gorm:"type:text;not null" json:"containers"`
}

// AppDaemonSet record created daemon set
type AppDaemonSet struct {
	ID            uint64    `gorm:"type:Integer;primaryKey;autoIncrement:true"`
	DaemonSetName string    `gorm:"type:char(128);unique;not null"`
	AppID         uint64    `gorm:"type:Integer;not null"`
	NodeGroupID   uint64    `gorm:"type:Integer;not null"`
	NodeGroupName string    `gorm:"type:char(64);not null"`
	CreatedAt     time.Time `gorm:"type:time"`
	UpdatedAt     time.Time `gorm:"type:time"`
}

// AppInstance is application instance
type AppInstance struct {
	ID             uint64    `gorm:"type:Integer;primaryKey;autoIncrement:true"`
	PodName        string    `gorm:"type:char(128);unique;not null"`
	NodeID         uint64    `gorm:"type:Integer;not null"`
	NodeName       string    `gorm:"type:char(64);not null"`
	NodeUniqueName string    `gorm:"type:char(64);not null"`
	NodeGroupID    uint64    `gorm:"type:Integer;not null"`
	AppID          uint64    `gorm:"type:Integer;not null"`
	AppName        string    `gorm:"type:char(128);not null"`
	CreatedAt      time.Time `gorm:"type:time"`
	UpdatedAt      time.Time `gorm:"type:time"`
	ContainerInfo  string    `gorm:"type:text;" json:"containers"`
}

// AppTemplateDb db table for application template group
type AppTemplateDb struct {
	ID           uint64    `gorm:"type:integer;primaryKey;autoIncrement:true"`
	TemplateName string    `gorm:"type:char(128);unique;not null"`
	Description  string    `gorm:"type:char(255);" json:"description"`
	CreatedAt    time.Time `gorm:"type:time"`
	UpdatedAt    time.Time `gorm:"type:time"`
	Containers   string    `gorm:"type:text;not null" json:"containers"`
}

// ConfigmapInfo is configmap table info
type ConfigmapInfo struct {
	ConfigmapID      int64     `gorm:"type:integer;primaryKey;autoIncrement:true"`
	ConfigmapName    string    `gorm:"type:char(64);unique;not null"`
	ConfigmapContent string    `gorm:"type:char(65535)"`
	Description      string    `gorm:"type:char(255)"`
	CreatedAt        time.Time `gorm:"type:time"`
	UpdatedAt        time.Time `gorm:"type:time"`
}
