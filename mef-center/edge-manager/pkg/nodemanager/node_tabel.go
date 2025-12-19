// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package nodemanager to init node database table
package nodemanager

// NodeInfo is node db table info
type NodeInfo struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement:true"  json:"id"`
	Description  string `gorm:"size:512"                       json:"description"`
	NodeType     string `gorm:"size:128"                       json:"nodeType"`
	NodeName     string `gorm:"size:255;unique;not null"       json:"nodeName"`
	UniqueName   string `gorm:"size:255;unique;not null"       json:"uniqueName"`
	SerialNumber string `gorm:"size:255;unique;not null"       json:"serialNumber"`
	IP           string `gorm:"size:256"                       json:"ip"`
	IsManaged    bool   `gorm:"size:4;not null"                json:"isManaged"`
	SoftwareInfo string `gorm:"type:text;not null"             json:"softwareInfo"`
	CreatedAt    string `gorm:"not null"                       json:"createdAt"`
	UpdatedAt    string `gorm:"not null"                       json:"updatedAt"`
}

// NodeGroup is node group db table
type NodeGroup struct {
	ID               uint64 `gorm:"primaryKey;autoIncrement:true"  json:"id"`
	Description      string `gorm:"size:512"                       json:"description"`
	GroupName        string `gorm:"size:255;unique;not null"       json:"groupName"`
	CreatedAt        string `gorm:"not null"                       json:"createdAt"`
	UpdatedAt        string `gorm:"not null"                       json:"updatedAt"`
	ResourcesRequest string `gorm:"type:text"                      json:"-"`
}

// NodeRelation is node relation table
type NodeRelation struct {
	GroupID   uint64 `gorm:"uniqueIndex:unique_relation;not null"`
	NodeID    uint64 `gorm:"uniqueIndex:unique_relation;not null"`
	CreatedAt string `gorm:"not null"`
}
