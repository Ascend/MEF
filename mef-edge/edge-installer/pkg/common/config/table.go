// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
