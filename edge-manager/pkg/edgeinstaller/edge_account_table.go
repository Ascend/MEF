// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller for table edge_account_infos definition
package edgeinstaller

// EdgeAccountInfo is edge account table info
type EdgeAccountInfo struct {
	ID        uint   `gorm:"primary_key;auto_increment"`
	Account   string `gorm:"size:256;unique"`
	Password  string `gorm:"size:256"`
	Salt      string `gorm:"size:32"`
	CreatedAt string `gorm:"not null"`
	UpdatedAt string `gorm:"not null"`
}
