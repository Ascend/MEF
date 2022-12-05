// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node database table
package nodemanager

// NodeInfo is node db table info
type NodeInfo struct {
	ID          int64  `gorm:"primaryKey;autoIncrement:true"  json:"id"`
	Description string `gorm:"size:256"                       json:"description"`
	NodeType    string `gorm:"size:128"                       json:"nodeType"`
	NodeName    string `gorm:"size:255;unique;not null"       json:"nodeName"`
	UniqueName  string `gorm:"size:255;unique;not null"       json:"uniqueName"`
	CPUCore     int64  `gorm:"size:32"                        json:"cpuCore"`
	Memory      int64  `gorm:"size:32"                        json:"memory"`
	NPUType     string `gorm:"size:256"                       json:"npuType"`
	NPUNum      int64  `gorm:"size:32"                        json:"npuNum"`
	Status      string `gorm:"size:255;not null"              json:"status"`
	IsManaged   bool   `gorm:"size:4;not null"                json:"isManaged"`
	CreatedAt   string `gorm:"not null"                       json:"createdAt"`
	UpdateAt    string `gorm:"not null"                       json:"updateAt"`
}

// NodeGroup is node group db table
type NodeGroup struct {
	ID          int64  `gorm:"primaryKey;autoIncrement:true"  json:"id"`
	Description string `gorm:"size:256"                       json:"description"`
	GroupName   string `gorm:"size:255;unique;not null"       json:"groupName"`
	CreatedAt   string `gorm:"not null"                       json:"createdAt"`
	UpdateAt    string `gorm:"not null"                       json:"updateAt"`
}

// NodeRelation is node relation table
type NodeRelation struct {
	GroupID   int64  `gorm:""`
	NodeID    int64  `gorm:""`
	CreatedAt string `gorm:"not null"`
}
