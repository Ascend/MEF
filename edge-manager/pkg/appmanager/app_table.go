// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager database table
package appmanager

// AppInfo is app db table info
type AppInfo struct {
	ID          uint64 `gorm:"type:integer;primaryKey;autoIncrement:true"`
	AppName     string `gorm:"type:char(128);unique;not null"`
	Description string `gorm:"type:char(255);" json:"description"`
	AppGroupID  uint64 `gorm:"type:integer;not null"`
	CreatedAt   string `gorm:"type:char(19);not null"`
	ModifiedAt  string `gorm:"type:char(19);not null"`
}

// AppContainer containers belonging to the same App share the same AppName.
type AppContainer struct {
	ID            uint64 `gorm:"type:integer;primaryKey;autoIncrement:true"`
	AppName       string `gorm:"type:varchar(32);not null"`
	CreatedAt     string `gorm:"type:char(19);not null"`
	ModifiedAt    string `gorm:"type:char(19);not null"`
	ContainerName string `gorm:"type:varchar(32);not null"`
	ImageName     string `gorm:"type:varchar(64);not null"`
	ImageVersion  string `gorm:"type:varchar(16);not null"`
	CpuRequest    string `gorm:"type:varchar(7);not null"`
	CpuLimit      string `gorm:"type:varchar(7)"`
	MemoryRequest string `gorm:"type:varchar(7);not null"`
	MemoryLimit   string `gorm:"type:varchar(7)"`
	Npu           string `gorm:"type:varchar(5)"`
	Env           string `gorm:"type:text"`
	UserID        int    `gorm:"type:integer"`
	GroupID       int    `gorm:"type:integer"`
	ContainerPort string `gorm:"type:varchar(5)"`
	HostIp        string `gorm:"type:varchar(15)"`
	HostPort      int    `gorm:"type:integer"`
	Command       string `gorm:"type:text"`
}

// AppInstance is application instance
type AppInstance struct {
	ID          int64  `gorm:"type:Integer;primaryKey;autoIncrement:true"`
	PodName     string `gorm:"type:char(42);unique;not null"`
	NodeName    string `gorm:"type:char(64);not null"`
	NodeGroupID int64  `gorm:"type:Integer;not null"`
	Status      string `gorm:"type:char(50);not null"`
	VersionID   string `gorm:"not null"`
	CreatedAt   string `gorm:"type:time;not null"`
	ChangedAt   string `gorm:"type:time;not null"`
}
