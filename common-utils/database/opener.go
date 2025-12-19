// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package database
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
)

const (
	dbPathMode         = 0600
	maxOpenConnections = 1
	maxDbSize          = 1 * 1024 * 1024 * 1024
)

type dbOpener struct {
	singletonMode bool
	mainDbPath    string
	Options

	backupPlugin *backuputils.DbBackupPlugin
	mainDb       *gorm.DB
}

func (d *dbOpener) open() (*gorm.DB, error) {
	if err := d.checkAndModifyArgs(); err != nil {
		return nil, err
	}

	if !d.EnableBackup {
		return openDbConnection(d.mainDbPath)
	}

	if err := d.setupBackupPlugin(); err != nil {
		return nil, err
	}

	if d.EnableAutoRecover {
		go d.runAutoRecoverLoop()
	}

	return d.mainDb, nil
}

func (d *dbOpener) checkAndModifyArgs() error {
	checkAndModifyFns := []func() error{
		d.checkAndModifyMainDbPath,
		d.checkAndModifyBackupArgs,
		d.checkAndModifyRecoverArgs,
	}
	for _, fn := range checkAndModifyFns {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}

func (d *dbOpener) runAutoRecoverLoop() {
	ctx := context.Background()
	for {
		select {
		case <-d.backupPlugin.CloseNotify():
		case <-ctx.Done():
			return
		}

		hwlog.RunLog.Error("db is broken, start to recover")
		var (
			err      error
			firstRun = true
		)
		for err != nil || firstRun {
			firstRun = false
			if err = d.setupBackupPlugin(); err != nil {
				hwlog.RunLog.Errorf("recover failed, %v", err)
				time.Sleep(d.TestInterval)
			}
		}

		hwlog.RunLog.Info("recover success")
		if d.RecoverCallback != nil {
			d.RecoverCallback(d.mainDb)
		}
	}
}

func (d *dbOpener) setupBackupPlugin() error {
	if err := backuputils.BackupOrRestoreDb(d.mainDbPath, d.BackupDbPath, openDbConnection); err != nil {
		return err
	}

	mainDb, err := openDbConnection(d.mainDbPath)
	if err != nil {
		return err
	}

	plugin := &backuputils.DbBackupPlugin{
		MainDbPath:   d.mainDbPath,
		BackupDbPath: d.BackupDbPath,
		Opener:       openDbConnection,
		TestInterval: d.TestInterval,
	}
	if err = mainDb.Use(plugin); err != nil {
		closeDbConnection(mainDb)
		return err
	}

	d.backupPlugin = plugin
	d.mainDb = mainDb
	return nil
}

func (d *dbOpener) checkAndModifyMainDbPath() error {
	mainDbPath, err := checkDbPath(d.mainDbPath)
	if err != nil {
		return errors.New("check db path error")
	}
	d.mainDbPath = mainDbPath
	return nil
}

func (d *dbOpener) checkAndModifyBackupArgs() error {
	if !d.EnableBackup {
		if !(d.BackupDbPath == "" &&
			d.TestInterval == 0 &&
			!d.EnableAutoRecover &&
			d.RecoverCallback == nil) {
			return errors.New("invalid backup options")
		}
		return nil
	}

	backupDbPath, err := checkDbPath(d.BackupDbPath)
	if err != nil {
		return err
	}
	d.BackupDbPath = backupDbPath
	if d.TestInterval <= 0 {
		return errors.New("invalid test interval")
	}
	return nil
}

func (d *dbOpener) checkAndModifyRecoverArgs() error {
	if !d.EnableAutoRecover {
		if d.RecoverCallback != nil {
			return errors.New("invalid recover options")
		}
		return nil
	}

	if d.singletonMode {
		if d.RecoverCallback != nil {
			return errors.New("recover callback is not allowed in singleton mode")
		}
		d.RecoverCallback = setDbSingleton
	}
	return nil
}

func checkDbPath(dbPath string) (string, error) {
	if dbPath == "" {
		return "", errors.New("invalid db path")
	}
	return fileutils.CheckOriginPath(dbPath)
}

func openDbConnection(dbPath string) (*gorm.DB, error) {
	var db *gorm.DB
	var sqlDB *sql.DB
	var err error
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		hwlog.RunLog.Errorf("initialize db session failed: %v", err)
		return nil, errors.New("initialize db session failed")
	}

	if sqlDB, err = db.DB(); sqlDB != nil {
		sqlDB.SetMaxOpenConns(maxOpenConnections)
	} else {
		hwlog.RunLog.Warn("set max open connections failed")
	}

	sizeChecker := fileutils.NewFileSizeChecker(maxDbSize)
	linkChecker := fileutils.NewFileLinkChecker(false)
	ownerChecker := fileutils.NewFileOwnerChecker(false, true, fileutils.RootUid, fileutils.RootGid)
	sizeChecker.SetNext(linkChecker)
	linkChecker.SetNext(ownerChecker)
	if err = fileutils.SetPathPermission(dbPath, dbPathMode, false, false, sizeChecker); err != nil {
		closeDbConnection(db)
		hwlog.RunLog.Errorf("set db file mode failed: %v", err)
		return nil, fmt.Errorf("set db file mode failed: %v", err)
	}

	return db, nil
}

func closeDbConnection(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		hwlog.RunLog.Error("failed to close db: failed to get sql db")
		return
	}
	if err := sqlDB.Close(); err != nil {
		hwlog.RunLog.Error("failed to close db: db error")
	}
}
