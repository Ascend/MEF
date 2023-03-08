// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package softwaremanager for db table
package softwaremanager

// SoftwareInfo software repo info
type SoftwareInfo struct {
	Key       string `gorm:"primaryKey"`
	Value     string `gorm:""`
	CreatedAt string `gorm:"not null"`
	UpdatedAt string `gorm:"not null"`
}
