// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//
//	http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package backuputils
package backuputils

import (
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
)

const (
	backupDbName = "backupDb"
	backupTxName = "backupTx"
)

// BackupOrRestoreDb backup or restore db
func BackupOrRestoreDb(mainDbPath, backupDbPath string, dbOpener DbOpener) error {
	if dbOpener == nil {
		return errors.New("opener is nil")
	}
	if !fileutils.IsExist(mainDbPath) && !fileutils.IsExist(backupDbPath) {
		hwlog.RunLog.Info("both main and backup db do not exist, skip recover")
		return nil
	}

	mainDbErr := testDbIntegrity(mainDbPath, dbOpener)
	backupDbErr := testDbIntegrity(backupDbPath, dbOpener)
	if mainDbErr != nil && backupDbErr != nil {
		hwlog.RunLog.Errorf("main db is broken, %v", mainDbErr)
		hwlog.RunLog.Errorf("backup db is broken, %v", backupDbErr)
		hwlog.RunLog.Error("both main and backup db are broken, unable to recover")
		return errors.New("both main and backup db are broken")
	}

	// restore
	if mainDbErr != nil {
		hwlog.RunLog.Errorf("main db is broken, %v", mainDbErr)
		if err := fileutils.CopyFile(backupDbPath, mainDbPath); err != nil {
			hwlog.RunLog.Errorf("restore db failed, %v", err)
			return errors.New("restore db failed")
		}
		hwlog.RunLog.Info("restore db success")
		return nil
	}

	// backup
	if backupDbErr != nil {
		hwlog.RunLog.Errorf("backup db is broken, %v", backupDbErr)
		if err := fileutils.CopyFile(mainDbPath, backupDbPath); err != nil {
			hwlog.RunLog.Errorf("backup db failed, %v", err)
			return errors.New("backup db failed")
		}
		hwlog.RunLog.Info("backup db success")
		return nil
	}

	hwlog.RunLog.Info("both main and backup db are valid, no need to recover")
	return nil
}

// Transaction runs fc in transaction
func Transaction(mainTx *gorm.DB, fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	if mainTx == nil || fc == nil {
		return errors.New("invalid transaction arguments")
	}
	backupTx, ok := getBackupTx(mainTx)
	if !ok {
		return mainTx.Transaction(fc, opts...)
	}

	if isTransaction(backupTx.Statement.ConnPool) {
		if !backupTx.DisableNestedTransaction {
			return runFuncWithNestedTransaction(mainTx, backupTx, fc)
		} else {
			return runFuncWithoutTransaction(mainTx, backupTx, fc)
		}
	}

	return runFuncWithTransaction(mainTx, backupTx, fc, opts...)
}

// GetBackupDb returns the backup db if exists
func GetBackupDb(db *gorm.DB) (*gorm.DB, bool) {
	if db == nil {
		return nil, false
	}
	backupDbVal := db.Statement.Context.Value(backupDbName)
	if backupDbVal == nil {
		return nil, false
	}
	backupDb, ok := backupDbVal.(*gorm.DB)
	return backupDb, ok
}

// testDbIntegrity tests integrity of db. Raise error if db does not exist
func testDbIntegrity(path string, opener DbOpener) error {
	if !fileutils.IsExist(path) {
		return errors.New("db is deleted")
	}

	db, err := opener(path)
	if err != nil {
		return errors.New("open db failed")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return errors.New("get sql db failed")
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			hwlog.RunLog.Error("failed to close sqlDB")
		}
	}()

	var result string
	stmt := db.Raw("PRAGMA integrity_check;").Scan(&result)
	if stmt.Error != nil {
		return errors.New("sql error")
	}
	// sqlite response ok if integrity_check success
	if result != "ok" {
		return errors.New("db is broken")
	}
	return nil
}

func runFuncWithoutTransaction(mainTx, backupTx *gorm.DB, fc func(*gorm.DB) error) error {
	var err error
	func() {
		defer func() {
			var panickedErr error
			if data := recover(); data != nil {
				panickedErr = fmt.Errorf("panicked, %v", data)
			}
			if err == nil && panickedErr != nil {
				err = panickedErr
			} else if err != nil && panickedErr != nil {
				hwlog.RunLog.Error("a panic error was discard")
			}
		}()
		mainTx = mainTx.Set(backupTxName, backupTx.Session(&gorm.Session{}))
		err = fc(mainTx)
	}()
	return err
}

func runFuncWithNestedTransaction(mainTx *gorm.DB, backupTx *gorm.DB, fc func(*gorm.DB) error) error {
	savePointName := fmt.Sprintf("sp%p", fc)
	if err := backupTx.SavePoint(savePointName).Error; err != nil {
		return err
	}

	var err error
	func() {
		defer func() {
			var panickedErr error
			if data := recover(); data != nil {
				panickedErr = fmt.Errorf("panicked, %v", data)
			}
			if err == nil && panickedErr != nil {
				err = panickedErr
			} else if err != nil && panickedErr != nil {
				hwlog.RunLog.Error("a panic error was discard")
			}
			if err != nil {
				backupTx.RollbackTo(savePointName)
			}
		}()
		mainTx = mainTx.Set(backupTxName, backupTx.Session(&gorm.Session{}))
		err = mainTx.Transaction(fc)
	}()
	return err
}

func runFuncWithTransaction(
	mainTx *gorm.DB, backupTx *gorm.DB, fc func(*gorm.DB) error, opts ...*sql.TxOptions) error {
	tx := backupTx.Begin(opts...)
	if tx.Error != nil {
		return tx.Error
	}

	var err error
	func() {
		defer func() {
			var panickedErr error
			if panickedData := recover(); panickedData != nil {
				panickedErr = fmt.Errorf("panicked, %v", panickedData)
			}
			if err == nil && panickedErr != nil {
				hwlog.RunLog.Error("a transaction error is overwrite because panic")
				err = panickedErr
			} else if err != nil && panickedErr != nil {
				hwlog.RunLog.Error("a panic error was discard")
			}
			if err != nil {
				tx.Rollback()
				return
			}
			tx.Commit()
		}()
		mainTx = mainTx.Set(backupTxName, tx)
		err = mainTx.Transaction(fc, opts...)
	}()
	return err
}

func getBackupTx(db *gorm.DB) (*gorm.DB, bool) {
	backupTxVal, ok := db.Get(backupTxName)
	if ok {
		backupTx, ok := backupTxVal.(*gorm.DB)
		if ok {
			return backupTx, true
		}
	}

	backupDbVal := db.Statement.Context.Value(backupDbName)
	if backupDbVal == nil {
		return nil, false
	}
	backupDb, ok := backupDbVal.(*gorm.DB)
	if !ok {
		return nil, false
	}
	session := backupDb.Session(&gorm.Session{})
	db.Statement.Settings.Store(backupTxName, session)
	return session, true
}

func isTransaction(connPool gorm.ConnPool) bool {
	committer, ok := connPool.(gorm.TxCommitter)
	if !ok {
		return false
	}
	return committer != nil
}
