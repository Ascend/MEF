// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeconnector used table
package edgeconnector

// ConnInfo is websocket connection db table info
type ConnInfo struct {
	ID        uint   `gorm:"primary_key"`
	Address   string `gorm:"size:16"`
	Port      string `gorm:"size:16"`
	Username  string `gorm:"size:32;unique"`
	Password  []byte `gorm:"size:32"`
	CreatedAt string `gorm:"not null"`
	UpdatedAt string `gorm:"not null"`
}
