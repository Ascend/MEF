// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package common for database init function
package common

import (
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"huawei.com/mindx/common/hwlog"
)

const (
	dBFileMode         = 0640
	maxOpenConnections = 1
)

// InitDbConnection init database connection
func InitDbConnection(dbPath string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		hwlog.RunLog.Error("init database connection failed")
		return nil
	}
	if err = os.Chmod(dbPath, dBFileMode); err != nil {
		hwlog.RunLog.Error("chmod for database file error")
		return nil
	}
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(maxOpenConnections)
	}
	hwlog.RunLog.Info("init database connection success")
	return db
}
