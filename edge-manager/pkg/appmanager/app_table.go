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
	Status      string `gorm:"type:char(128)"`
	Containers  string `gorm:"type:text;not null" json:"containers"`
}

// AppContainer containers belonging to the same App share the same AppName.
type AppContainer struct {
	ID            uint64 `gorm:"type:integer;primaryKey;autoIncrement:true" json:"id"`
	AppName       string `gorm:"type:varchar(32)" json:"appName"`
	CreatedAt     string `gorm:"type:char(19);not null" json:"createdAt"`
	ModifiedAt    string `gorm:"type:char(19);not null" json:"modifiedAt"`
	ContainerName string `gorm:"type:varchar(32);not null" json:"containerName"`
	ImageName     string `gorm:"type:varchar(64);not null" json:"imageName"`
	ImageVersion  string `gorm:"type:varchar(16);not null" json:"imageVersion"`
	CpuRequest    string `gorm:"type:varchar(7);not null" json:"cpuRequest"`
	CpuLimit      string `gorm:"type:varchar(7)" json:"cpuLimit"`
	MemoryRequest string `gorm:"type:varchar(7);not null" json:"memoryRequest"`
	MemoryLimit   string `gorm:"type:varchar(7)" json:"memoryLimit"`
	Npu           string `gorm:"type:varchar(5)" json:"npu"`
	Env           string `gorm:"type:text" json:"env"`
	UserID        int    `gorm:"type:integer" json:"userID"`
	GroupID       int    `gorm:"type:integer" json:"groupID"`
	ContainerPort string `gorm:"type:varchar(5)" json:"containerPort"`
	HostIp        string `gorm:"type:varchar(15)" json:"hostIp"`
	HostPort      int    `gorm:"type:integer" json:"hostPort"`
	Command       string `gorm:"type:text" json:"command"`
}

// AppInstance is application instance
type AppInstance struct {
	ID            int64  `gorm:"type:Integer;primaryKey;autoIncrement:true"`
	PodName       string `gorm:"type:char(42);unique;not null"`
	NodeName      string `gorm:"type:char(64);not null"`
	NodeGroupName string `gorm:"type:char(64);not null"`
	NodeGroupID   int64  `gorm:"type:Integer;not null"`
	Status        string `gorm:"type:char(50);not null"`
	AppID         string `gorm:"not null"`
	CreatedAt     string `gorm:"type:time;not null"`
	ChangedAt     string `gorm:"type:time;not null"`
}
