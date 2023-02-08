// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package usermgr to init user manager database table
package usermgr

// User the user table struct
type User struct {
	ID                 uint64 `gorm:"primaryKey;autoIncrement:true" json:"userid"`
	Username           string `gorm:"size:32;unique;not null" json:"username"`
	Password           string `gorm:"size:256;not null" json:"-"`
	Salt               string `gorm:"size:32;not null" json:"-"`
	LoginTime          string
	LoginFailTime      string `json:"-"`
	PasswordModifyTime string `json:"-"`
	PasswordWrongTimes uint64 `gorm:"type:Integer;default:0;not null" json:"-"`
	FirstLogin         bool   `gorm:"size:4;default:true;not null" json:"isFirstLogin"`
	LockState          bool   `gorm:"size:4;default:false;not null" json:"-"`
	LockTime           string `json:"-"`
	CreatedAt          string `gorm:"not null" json:"-"`
	UpdatedAt          string `json:"-"`
}

// IpForbidden the struct for table IpForbidden
type IpForbidden struct {
	ID       uint64 `gorm:"primaryKey;autoIncrement:true"`
	Ip       string `gorm:"size:128;unique;not null"`
	LockTime string
}

// HistoryPassword the struct for table history_passowrd
type HistoryPassword struct {
	ID              uint64 `gorm:"primaryKey;autoIncrement:true"`
	UserId          uint64 `gorm:"not null"`
	HistoryPassword string `gorm:"size:256;not null"`
	Salt            string `gorm:"size:32;not null"`
	CreatedAt       string
}
