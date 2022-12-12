// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package database to init database
package database

import (
	"errors"
	"flag"
	"fmt"

	"gorm.io/gorm"
	"huawei.com/mindxedge/base/common"
)

const (
	// DBPATH sqlite database path
	DBPATH = "/etc/mindx-edge/edge-manager/edge-manager.db"
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
	gormDB = common.InitDbConnection(dbPath)
	if gormDB == nil {
		return fmt.Errorf("initialise database failed")
	}
	return nil
}

// GetDb connection data
func GetDb() *gorm.DB {
	return gormDB
}

// CreateTableIfNotExists create table
func CreateTableIfNotExists(modelType interface{}) error {
	if modelType == nil || gormDB == nil {
		return errors.New("create table failed")
	}
	return gormDB.AutoMigrate(modelType)
}

// GetItemCount get item count in table
func GetItemCount(table interface{}) (int, error) {
	var total int64
	if err := GetDb().Model(table).Count(&total).Error; err != nil {
		return 0, err
	}
	return int(total), nil
}
