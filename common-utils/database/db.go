// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package database for
package database

import (
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/hwlog"
)

var gormDB *gorm.DB

// Options database  options
type Options struct {
	EnableBackup      bool
	BackupDbPath      string
	TestInterval      time.Duration
	EnableAutoRecover bool
	RecoverCallback   func(*gorm.DB)
}

// InitDB init database, dir and file mode is constant dbPathMode.
func InitDB(dbPath string, opts ...Options) error {
	if len(opts) == 0 {
		opts = append(opts, Options{})
	}
	opener := dbOpener{
		singletonMode: true,
		mainDbPath:    dbPath,
		Options:       opts[0],
	}

	db, err := opener.open()
	if err != nil {
		hwlog.RunLog.Errorf("initialize db connection failed, %v", err)
		return err
	}
	hwlog.RunLog.Info("initialize db connection success")
	setDbSingleton(db)
	return nil
}

// InitDbConnection init database connection
func InitDbConnection(dbPath string, opts ...Options) *gorm.DB {
	if len(opts) == 0 {
		opts = append(opts, Options{})
	}
	opener := dbOpener{
		mainDbPath: dbPath,
		Options:    opts[0],
	}

	db, err := opener.open()
	if err != nil {
		hwlog.RunLog.Errorf("initialize db connection failed, %v", err)
		return nil
	}
	hwlog.RunLog.Info("initialize db connection success")
	return db
}

// GetDb get database
func GetDb() *gorm.DB {
	return gormDB
}

// CreateTableIfNotExist create table if it does not exist
func CreateTableIfNotExist(modelType interface{}) error {
	mainDb := gormDB
	if modelType == nil || mainDb == nil {
		return errors.New("create table failed")
	}

	if err := mainDb.AutoMigrate(modelType); err != nil {
		return err
	}
	backupDb, ok := backuputils.GetBackupDb(mainDb)
	if !ok {
		return nil
	}
	return backupDb.AutoMigrate(modelType)
}

// DropTableIfExist drop table if it exists
func DropTableIfExist(modelType interface{}) error {
	mainDb := gormDB
	if modelType == nil || mainDb == nil {
		return errors.New("drop table failed")
	}

	if err := dropTableIfExist(mainDb, modelType); err != nil {
		return err
	}
	backupDb, ok := backuputils.GetBackupDb(mainDb)
	if !ok {
		return nil
	}
	return dropTableIfExist(backupDb, modelType)
}

// Transaction transaction util
func Transaction(db *gorm.DB, fc func(*gorm.DB) error, opts ...*sql.TxOptions) error {
	return backuputils.Transaction(db, fc, opts...)
}

func dropTableIfExist(db *gorm.DB, modelType interface{}) error {
	if db.Migrator().HasTable(modelType) {
		return db.Migrator().DropTable(modelType)
	}
	return nil
}

func setDbSingleton(db *gorm.DB) {
	gormDB = db
}
