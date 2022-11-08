// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package database to init database
package database

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"huawei.com/mindx/common/hwlog"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	// DBPATH sqlite database path
	DBPATH = "/etc/mindx-edge/edge-manager/edge-manager.db"
	// DBFileMode database file mode
	DBFileMode         = 0640
	maxOpenConnections = 1
)

var (
	dbPath string
	gormDB *gorm.DB
)

func init() {
	// database path configuration
	flag.StringVar(&dbPath, "dbPath", DBPATH, "sqlite database path")
}

// InitDB init database client
func InitDB() error {
	initDbConnection()
	if gormDB == nil {
		return fmt.Errorf("initialise database failed")
	}
	return nil
}

// GetDb connection data
func GetDb() *gorm.DB {
	return gormDB
}

// InitDbConnection init database connection
func initDbConnection() {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		hwlog.RunLog.Error("init database connection failed")
		return
	}
	if err = os.Chmod(dbPath, DBFileMode); err != nil {
		hwlog.RunLog.Error("chmod for database file error")
		return
	}
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(maxOpenConnections)
	}
	gormDB = db
	hwlog.RunLog.Info("init database connection success")
}

// CreateTableIfNotExists create table
func CreateTableIfNotExists(modelType interface{}) error {
	if modelType == nil || gormDB == nil {
		return errors.New("create table failed")
	}
	return gormDB.AutoMigrate(modelType)
}
