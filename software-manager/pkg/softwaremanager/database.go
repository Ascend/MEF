// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

package softwaremanager

import (
	"errors"
	"fmt"
	"path/filepath"

	"gorm.io/gorm"
	"huawei.com/mindxedge/base/common"
)

var gormDB *gorm.DB

// InitDB is used to init database in main.go
func InitDB() error {
	gormDB = common.InitDbConnection(filepath.Join(RepositoryFilesPath, "/repository.db"))
	if gormDB == nil {
		return fmt.Errorf("initialise database failed")
	}
	if err := createTableIfNotExists(&softwareRecord{}); err != nil {
		return fmt.Errorf("initialise table failed")
	}

	return nil
}

func getDb() *gorm.DB {
	return gormDB
}

func createTableIfNotExists(modelType interface{}) error {
	if modelType == nil || gormDB == nil {
		return errors.New("create table failed")
	}
	return gormDB.AutoMigrate(modelType)
}
