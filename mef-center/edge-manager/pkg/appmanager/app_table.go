// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package appmanager to init app manager database table
package appmanager

import "time"

// AppInfo is app db table info
type AppInfo struct {
	ID          uint64 `gorm:"type:integer;primaryKey;autoIncrement:true"`
	AppName     string `gorm:"type:char(128);unique;not null"`
	Description string `gorm:"type:char(255);" json:"description"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Containers  string `gorm:"type:text;not null" json:"containers"`
}

// AppDaemonSet record created daemon set
// property NodeGroupName is deprecated
type AppDaemonSet struct {
	ID            uint64 `gorm:"type:Integer;primaryKey;autoIncrement:true"`
	DaemonSetName string `gorm:"type:char(128);unique;not null"`
	AppID         uint64 `gorm:"type:Integer;not null"`
	NodeGroupID   uint64 `gorm:"type:Integer;not null"`
	NodeGroupName string `gorm:"type:char(64);not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// AppInstance is application instance
type AppInstance struct {
	ID             uint64 `gorm:"type:Integer;primaryKey;autoIncrement:true"`
	PodName        string `gorm:"type:char(128);unique;not null"`
	NodeID         uint64 `gorm:"type:Integer;not null"`
	NodeName       string `gorm:"type:char(64);not null"`
	NodeUniqueName string `gorm:"type:char(64);not null"`
	NodeGroupID    uint64 `gorm:"type:Integer;not null"`
	AppID          uint64 `gorm:"type:Integer;not null"`
	AppName        string `gorm:"type:char(128);not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ContainerInfo  string `gorm:"type:text;" json:"containers"`
}
