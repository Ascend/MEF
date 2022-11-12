// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node database table
package nodemanager

// NodeInfo is node db table info
type NodeInfo struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement:true"`
	Description string `gorm:"size:256" json:"description"`
	NodeType    string `gorm:"size:128"`
	NodeName    string `gorm:"size:255;unique;not null"`
	UniqueName  string `gorm:"size:255;unique;not null"`
	CPUCore     uint64 `gorm:"size:32"`
	Memory      uint64 `gorm:"size:32"`
	NPUType     string `gorm:"size:256"`
	NPUNum      uint64 `gorm:"size:32"`
	Status      string `gorm:"size:255;not null"`
	IsManaged   bool   `gorm:"size:4;not null"`
	CreatedAt   string `gorm:"not null"`
	UpdateAt    string `gorm:"not null"`
}

// NodeGroup is node group db table
type NodeGroup struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement:true"`
	Description string `gorm:"size:256"`
	GroupName   string `gorm:"size:255;unique;not null"`
	Label       string `gorm:"size:255;unique;not null"`
	CreatedAt   string `gorm:"not null"`
	UpdateAt    string `gorm:"not null"`
}

// NodeRelation is node relation table
type NodeRelation struct {
	GroupID   uint64 `gorm:""`
	NodeID    uint64 `gorm:""`
	CreatedAt string `gorm:"not null"`
}
