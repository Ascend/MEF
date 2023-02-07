// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package database to init database
package database

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"
)

var gormDB *gorm.DB

// InitDB init database client
func InitDB(dbPath string) error {
	err := utils.MakeSureDir(dbPath)
	if err != nil {
		return err
	}
	gormDB = common.InitDbConnection(dbPath)
	if gormDB == nil {
		return fmt.Errorf("initialize database failed")
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
	return gormDB.Debug().AutoMigrate(modelType)
}
