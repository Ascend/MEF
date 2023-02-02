// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgemsgmanager software manager info db table
package edgemsgmanager

// SoftwareMgrInfo software manager info db table
type SoftwareMgrInfo struct {
	ID        uint   `gorm:"primary_key"`
	Address   string `gorm:"size:16"`
	Port      string `gorm:"size:16"`
	Route     string `gorm:"size:32"`
	CreatedAt string `gorm:"not null"`
	UpdatedAt string `gorm:"not null"`
}
