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

var (
	gormDB *gorm.DB
)

// InitDB init database client
func InitDB(dbPath string) error {
	err := utils.MakeSureDir(dbPath)
	if err != nil {
		return err
	}
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
