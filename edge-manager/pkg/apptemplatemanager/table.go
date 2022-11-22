// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package apptemplatemanager to define database table struct
package apptemplatemanager

// AppTemplateDb db table for application template group
type AppTemplateDb struct {
	Id          uint64                `gorm:"column:Id;primary_key;not null;auto_increment"`
	Name        string                `gorm:"column:Name;size:64;unique;not null"`
	Description string                `gorm:"column:Description;size:255;not null"`
	CreatedAt   string                `gorm:"column:CreatedAt;size:19;not null"`
	ModifiedAt  string                `gorm:"column:ModifiedAt;size:19;not null"`
	Containers  []TemplateContainerDb `gorm:"-"`
}

// TemplateContainerDb db table for application container template
type TemplateContainerDb struct {
	Id             uint64 `gorm:"column:Id;primary_key;not null;auto_increment"`
	TemplateId     uint64 `gorm:"column:TemplateId;not null;uniqueIndex:idx_template_container"`
	Name           string `gorm:"column:Name;size:32;not null;uniqueIndex:idx_template_container"`
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
