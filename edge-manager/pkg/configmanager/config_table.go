// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package configmanager for
package configmanager

import "time"

// TokenInfo is token db table info
type TokenInfo struct {
	Token      []byte    `gorm:"size:32;not null"`
	Salt       []byte    `gorm:"size:16;not null"`
	ExpireTime time.Time `gorm:"size:255;not null"`
}
