// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package database to init database
package database

// AppTemplateGroupDb db table for application template group
type AppTemplateGroupDb struct {
	Id          uint64 `gorm:"column:Id;primary_key;not null;auto_increment"`
	Name        string `gorm:"column:Name;size:128;unique;not null"`
	Description string `gorm:"column:Description;size:255;not null"`
	CreatedAt   string `gorm:"column:CreatedAt;size:19;not null"`
	ModifiedAt  string `gorm:"column:ModifiedAt;size:19;not null"`
}

// AppContainerTemplateDb db table for application container template
type AppContainerTemplateDb struct {
	Id             uint64 `gorm:"column:Id;primary_key;not null;auto_increment"`
	GroupId        uint64 `gorm:"column:GroupId;not null"`
	VersionName    string `gorm:"column:VersionName;size:64;not null"`
	ContainerName  string `gorm:"column:ContainerName;size:32;not null"`
	CreatedAt      string `gorm:"column:CreatedAt;size:19;not null"`
	ModifiedAt     string `gorm:"column:ModifiedAt;size:19;not null"`
	ImageNme       string `gorm:"column:ImageNme;size:64;not null"`
	ImageVersion   string `gorm:"column:ImageVersion;size:16;not null"`
	CpuRequest     string `gorm:"column:CpuRequest;size:7;not null"`
	CpuLimit       string `gorm:"column:CpuLimit;size:7"`
	MemoryRequest  string `gorm:"column:MemoryRequest;size:7;not null"`
	MemoryLimit    string `gorm:"column:MemoryLimit;size:7"`
	Npu            string `gorm:"column:Npu;size:5"`
	Env            string `gorm:"column:Env"`
	ContainerUser  string `gorm:"column:ContainerUser;size:5"`
	ContainerGroup string `gorm:"column:ContainerGroup;size:5"`
	PortMaps       string `gorm:"column:PortMaps"`
	Command        string `gorm:"column:Command"`
}
