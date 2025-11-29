// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package config used table and struct
package config

// Configuration edge-installer config struct
type Configuration struct {
	Key       string `gorm:"primaryKey"`
	Value     []byte `gorm:""`
	CreatedAt string `gorm:"not null"`
	UpdatedAt string `gorm:"not null"`
}

// AlarmConfig alarm config table
type AlarmConfig struct {
	ConfigName  string `gorm:"primaryKey"`
	ConfigValue int    `gorm:"not null"`
	HasModified *bool  `gorm:"not null"`
}
