// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node database table
package nodemanager

// NodeInfo is node db table info
type NodeInfo struct {
	ID              uint64 `gorm:"primaryKey;autoIncrement:true"  json:"id"`
	Description     string `gorm:"size:256"                       json:"description"`
	NodeType        string `gorm:"size:128"                       json:"nodeType"`
	NodeName        string `gorm:"size:255;unique;not null"       json:"nodeName"`
	UniqueName      string `gorm:"size:255;unique;not null"       json:"uniqueName"`
	IP              string `gorm:"size:256"                       json:"ip"`
	IsManaged       bool   `gorm:"size:4;not null"                json:"isManaged"`
	SoftwareInfo    string `gorm:"type:text;not null"             json:"softwareInfo"`
	UpgradeProgress string `gorm:"size:128"                       json:"upgradeProgress"`
	CreatedAt       string `gorm:"not null"                       json:"createdAt"`
	UpdatedAt       string `gorm:"not null"                       json:"updatedAt"`
}

// NodeGroup is node group db table
type NodeGroup struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement:true"  json:"id"`
	Description string `gorm:"size:256"                       json:"description"`
	GroupName   string `gorm:"size:255;unique;not null"       json:"groupName"`
	CreatedAt   string `gorm:"not null"                       json:"createdAt"`
	UpdatedAt   string `gorm:"not null"                       json:"updatedAt"`
}

// NodeRelation is node relation table
type NodeRelation struct {
	GroupID   uint64 `gorm:"uniqueIndex:unique_relation;not null"`
	NodeID    uint64 `gorm:"uniqueIndex:unique_relation;not null"`
	CreatedAt string `gorm:"not null"`
}
