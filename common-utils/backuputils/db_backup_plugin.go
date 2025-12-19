// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package backuputils
package backuputils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"huawei.com/mindx/common/hwlog"
)

const (
	dbBackupPluginName = "dbBackupPlugin"
	backupTxInfoName   = "backupTxInfo"
)

// DbOpener opener for db
type DbOpener func(string) (*gorm.DB, error)

// DbBackupPlugin database backup plugin
type DbBackupPlugin struct {
	MainDbPath   string
	BackupDbPath string
	Opener       DbOpener
	TestInterval time.Duration

	mainDb      *gorm.DB
	backupDb    *gorm.DB
	closeNotify chan struct{}
}

// Name returns the name of plugin
func (p *DbBackupPlugin) Name() string {
	return dbBackupPluginName
}

// Initialize inits the plugin
func (p *DbBackupPlugin) Initialize(mainDb *gorm.DB) error {
	if p.Opener == nil || p.MainDbPath == "" || p.BackupDbPath == "" {
		return errors.New("invalid plugin argument")
	}

	backupDb, err := p.Opener(p.BackupDbPath)
	if err != nil {
		return fmt.Errorf("open db failed, %v", err)
	}
	if err := p.registerCallbacks(mainDb); err != nil {
		return fmt.Errorf("register callback failed, %v", err)
	}
	mainDb.Statement.Context = context.WithValue(mainDb.Statement.Context, backupDbName, backupDb)

	p.closeNotify = make(chan struct{})
	p.mainDb = mainDb
	p.backupDb = backupDb

	go p.houseKeeping(mainDb.Statement.Context)
	return nil
}

// CloseNotify returns a channel to notify close event
func (p *DbBackupPlugin) CloseNotify() <-chan struct{} {
	return p.closeNotify
}

func (p *DbBackupPlugin) registerCallbacks(db *gorm.DB) error {
	if err := db.Callback().Create().After("gorm:save_after_associations").
		Register("backup:save_after_associations", p.afterBindArgs); err != nil {
		return err
	}
	if err := db.Callback().Create().After("gorm:create").
		Register("backup:create", p.afterExec); err != nil {
		return err
	}

	if err := db.Callback().Update().After("gorm:save_before_associations").
		Register("backup:save_before_associations", p.afterBindArgs); err != nil {
		return err
	}
	if err := db.Callback().Update().After("gorm:update").
		Register("backup:update", p.afterExec); err != nil {
		return err
	}

	if err := db.Callback().Delete().After("gorm:delete_before_associations").
		Register("backup:delete_before_associations", p.afterBindArgs); err != nil {
		return err
	}
	if err := db.Callback().Delete().After("gorm:delete").
		Register("backup:delete", p.afterExec); err != nil {
		return err
	}

	return nil
}

func (p *DbBackupPlugin) afterBindArgs(tx *gorm.DB) {
	if tx.Error != nil || tx.DryRun {
		return
	}
	backupTx, ok := getBackupTx(tx)
	if !ok {
		return
	}

	var (
		txCommitter      *gorm.DB
		savePointSession *gorm.DB
		savePointName    string
	)
	// begin the transaction if current session is not within a transaction
	if !isTransaction(backupTx.Statement.ConnPool) {
		backupTx = backupTx.Begin()
		txCommitter = backupTx
	}

	// create save point
	session := backupTx.Session(&gorm.Session{})
	if !tx.Config.DisableNestedTransaction {
		savePointName = fmt.Sprintf("sp%p", session)
		savePointSession = session.SavePoint(savePointName)
		session = savePointSession
	}

	// redo the sql
	_, err := session.Statement.ConnPool.ExecContext(
		session.Statement.Context, tx.Statement.SQL.String(), tx.Statement.Vars...)

	// setup txInfo
	backupTxInfo := &txInfo{
		txCommitter:      txCommitter,
		savePointSession: savePointSession,
		savePointName:    savePointName,
	}
	if err == nil {
		tx.InstanceSet(backupTxInfoName, backupTxInfo)
		return
	}

	// rollback
	hwlog.RunLog.Error("execute sql on backup db failed, start to rollback")
	if err := backupTxInfo.rollback(); err != nil {
		hwlog.RunLog.Error("rollback backup db failed")
	} else {
		hwlog.RunLog.Info("rollback backup db success")
	}

	// tell main db to rollback
	if err := tx.Statement.AddError(err); err != nil {
	}
}

func (p *DbBackupPlugin) afterExec(tx *gorm.DB) {
	if tx.DryRun {
		return
	}
	backupTxInfoVal, ok := tx.InstanceGet(backupTxInfoName)
	if !ok {
		return
	}
	backupTxInfo, ok := backupTxInfoVal.(*txInfo)
	if !ok {
		return
	}

	// commit
	if tx.Error == nil {
		// rollback the main db if backup db failed to commit
		if err := backupTxInfo.commit(); err != nil {
			hwlog.RunLog.Error("execute sql on main db success, but backup db failed to commit")
			err = tx.AddError(err)
		}
		return
	}

	// rollback
	hwlog.RunLog.Error("execute sql on main db failed, try to rollback")
	if err := backupTxInfo.rollback(); err != nil {
		hwlog.RunLog.Error("rollback backup db failed")
	} else {
		hwlog.RunLog.Info("rollback backup db success")
	}
}

func (p *DbBackupPlugin) houseKeeping(ctx context.Context) {
	ticker := time.NewTicker(p.TestInterval)
	defer func() {
		ticker.Stop()
		p.closeDbConnections()
		close(p.closeNotify)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := testDbIntegrity(p.MainDbPath, p.Opener); err != nil {
				hwlog.RunLog.Errorf("check main db integrity failed, %v", err)
				return
			}
			if err := testDbIntegrity(p.BackupDbPath, p.Opener); err != nil {
				hwlog.RunLog.Errorf("check backup db integrity failed, %v", err)
				return
			}
		}
	}
}

func (p *DbBackupPlugin) closeDbConnections() {
	mainSqlDB, err := p.mainDb.DB()
	if err != nil {
		hwlog.RunLog.Warn("main db type error, can't get sql db")
	} else {
		if err := mainSqlDB.Close(); err != nil {
			hwlog.RunLog.Warn("failed to close db")
		} else {
			hwlog.RunLog.Info("close main db success")
		}
	}

	backupSqlDB, err := p.backupDb.DB()
	if err != nil {
		hwlog.RunLog.Warn("backup db type error, can't get sql db")
	} else {
		if err := backupSqlDB.Close(); err != nil {
			hwlog.RunLog.Warn("failed to close db")
		} else {
			hwlog.RunLog.Info("close backup db success")
		}
	}
}

type txInfo struct {
	txCommitter      *gorm.DB
	savePointSession *gorm.DB
	savePointName    string
}

func (t txInfo) commit() error {
	if t.txCommitter != nil {
		return t.txCommitter.Commit().Error
	}
	return nil
}

func (t txInfo) rollback() error {
	var lastErr error
	if t.savePointSession != nil {
		lastErr = t.savePointSession.RollbackTo(t.savePointName).Error
	}

	if t.txCommitter == nil {
		return lastErr
	}
	if err := t.txCommitter.Rollback().Error; err != nil {
		if lastErr != nil {
			hwlog.RunLog.Error("failed to rollback tx, an error was hide")
		}
		lastErr = err
	}
	return lastErr
}
