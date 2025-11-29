// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package util
package util

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
)

// StartBackupEdgeOmDb starts backup process for edge om
func StartBackupEdgeOmDb(ctx context.Context) (context.Context, error) {
	return startBackupDb(initEdgeOmDbBackupTask, ctx)
}

// StartBackupEdgeCoreDb starts backup process for edge core
func StartBackupEdgeCoreDb(ctx context.Context) (context.Context, error) {
	return startBackupDb(initEdgeCoreBackupTask, ctx)
}

// CheckEdgeDbIntegrity checks integrity of MEFEdge' database
func CheckEdgeDbIntegrity() error {
	taskInitFns := []taskInitFn{
		initEdgeOmDbBackupTask,
		initEdgeCoreBackupTask,
	}
	for _, fn := range taskInitFns {
		task, err := fn()
		if err != nil {
			return err
		}
		if err := task.checkMainDbIntegrity(); err != nil {
			return err
		}
	}
	return nil
}

const (
	defaultTaskInterval = 5 * time.Minute
)

type taskInitFn func() (dbBackupTask, error)

func startBackupDb(initFn taskInitFn, parentCtx context.Context) (context.Context, error) {
	task, err := initFn()
	if err != nil {
		hwlog.RunLog.Errorf("failed to create db backup task, %v", err)
		return nil, err
	}
	if _, err := task.run(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(parentCtx)
	go backupDbLoop(task, ctx, cancel)
	return ctx, nil
}

func backupDbLoop(task dbBackupTask, ctx context.Context, cancel context.CancelFunc) {
	ticker := time.NewTicker(defaultTaskInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-ticker.C:
			if !ok {
				cancel()
				return
			}
		}
		restored, err := task.run()
		if err != nil {
			hwlog.RunLog.Errorf(err.Error())
			continue
		}
		if restored {
			cancel()
			return
		}
	}
}

// NewEdgeOmDbBackupMgr creates DbBackupMgr for EdgeOm
func initEdgeOmDbBackupTask() (dbBackupTask, error) {
	configPathMgr, err := path.GetConfigPathMgr()
	if err != nil {
		return nil, fmt.Errorf("get config path manager failed, %v", err)
	}
	return dbBackupTask{{mainDbPath: configPathMgr.GetEdgeOmDbPath()}}, nil
}

// NewEdgeCoreDbBackupMgr creates DbBackupMgr for EdgeCore And EdgeMain
func initEdgeCoreBackupTask() (dbBackupTask, error) {
	uid, gid, err := GetMefId()
	if err != nil {
		return nil, fmt.Errorf("failed to get uid of %s, %v", constants.EdgeUserName, err)
	}
	configPathMgr, err := path.GetConfigPathMgr()
	if err != nil {
		return nil, fmt.Errorf("get config path manager failed, %v", err)
	}
	return dbBackupTask{
		{mainDbPath: configPathMgr.GetEdgeMainDbPath(), uid: uid, gid: gid},
		{mainDbPath: configPathMgr.GetEdgeCoreDbPath()},
	}, nil
}

// dbBackupTask checks a set of database and restore them all if any db is broken
type dbBackupTask []singleDbBackupTask

type singleDbBackupTask struct {
	mainDbPath string
	uid        uint32
	gid        uint32
}

func (t dbBackupTask) run() (bool, error) {
	if err := t.checkMainDbIntegrity(); err != nil {
		hwlog.RunLog.Error("check integrity failed for main db, start to restore db")
		return true, t.restore()
	}

	hwlog.RunLog.Info("check integrity for main db successful, start to backup db")
	if err := t.backup(); err != nil {
		hwlog.RunLog.Errorf("backup failed, %v", err)
	}
	return false, nil
}

func (t dbBackupTask) checkMainDbIntegrity() error {
	for _, task := range t {
		if err := task.checkIntegrity(task.mainDbPath); err != nil {
			hwlog.RunLog.Errorf("check integrity of %s failed, %v", task.getDbName(), err)
			return err
		}
	}
	return nil
}

func (t dbBackupTask) checkBackupDbIntegrity() error {
	for _, task := range t {
		if err := task.checkIntegrity(task.getBackupDbPath()); err != nil {
			hwlog.RunLog.Errorf("check integrity of %s failed, %v", task.getBackupDbName(), err)
			return err
		}
	}
	return nil
}

func (t dbBackupTask) restore() error {
	if err := t.checkBackupDbIntegrity(); err != nil {
		return errors.New("backup db is broken")
	}
	for _, task := range t {
		if err := task.restore(); err != nil {
			hwlog.RunLog.Errorf("restore %s failed, %v", task.getDbName(), err)
			return fmt.Errorf("restore %s failed", task.getDbName())
		}
	}
	hwlog.RunLog.Info("restore db successful")
	return nil
}

func (t dbBackupTask) backup() error {
	var success bool
	defer func() {
		if success {
			return
		}

		for _, task := range t {
			if err := fileutils.DeleteFile(task.getBackupDbPath()); err != nil {
				hwlog.RunLog.Errorf("remove %s failed, %v", task.getBackupDbName(), err)
			}
		}
	}()

	for _, task := range t {
		if err := task.backup(); err != nil {
			hwlog.RunLog.Errorf("backup %s failed, %v", task.getDbName(), err)
			return fmt.Errorf("backup %s failed", task.getDbName())
		}
	}
	success = true
	hwlog.RunLog.Info("backup db successful")
	return nil
}

const (
	commandSqlite3         = "sqlite3"
	commandWaitTimeSeconds = 10
	resultOk               = "ok"
	sqlIntegrityCheck      = "PRAGMA integrity_check;"
	sqliteBackupCommand    = ".backup '%s'"
	sqliteRestoreCommand   = ".restore '%s'"
)

func (t singleDbBackupTask) checkIntegrity(dbPath string) error {
	if !fileutils.IsExist(dbPath) {
		return errors.New("db is deleted")
	}

	if err := t.prepareSingleDb(dbPath); err != nil {
		return fmt.Errorf("failed to prepare db, %v", err)
	}
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return errors.New("failed to open db connection")
	}
	defer func() {
		sqlDB, err := db.DB()
		if err != nil {
			hwlog.RunLog.Error("failed to get sql db")
			return
		}
		if err := sqlDB.Close(); err != nil {
			hwlog.RunLog.Error("failed to close sql db")
		}
	}()

	var result string
	if err := db.Raw(sqlIntegrityCheck).Scan(&result).Error; err != nil {
		return errors.New("db check sql failed")
	}
	if result != resultOk {
		return errors.New("integrity check failed, response is not ok")
	}
	return nil
}

func (t singleDbBackupTask) gracefulRestore() error {
	result := t.executeSqlite3Command(t.mainDbPath, fmt.Sprintf(sqliteRestoreCommand, t.getBackupDbPath()))
	if result.Err != nil {
		return fmt.Errorf("exec restore error: exit status %d", result.ExitCode)
	}
	if err := t.checkIntegrity(t.mainDbPath); err != nil {
		return errors.New("integrity check failed")
	}
	return nil
}

func (t singleDbBackupTask) restore() error {
	if err := t.prepareDb(); err != nil {
		hwlog.RunLog.Errorf("failed to prepare db, %v", err)
		return err
	}
	const maxRetry = 3
	if err := retry(t.gracefulRestore, maxRetry); err != nil {
		hwlog.RunLog.Errorf("graceful restore %s failed, %v", t.getDbName(), err)
		copyFn := func() error { return t.forciblyCopy(t.getBackupDbPath(), t.mainDbPath) }
		if err := retry(copyFn, maxRetry); err != nil {
			hwlog.RunLog.Errorf("forcibly restore %s failed, %v", t.getDbName(), err)
			return err
		}
	}
	hwlog.RunLog.Infof("restore %s successful", t.getDbName())
	return nil
}

func (t singleDbBackupTask) gracefulBackup() error {
	result := t.executeSqlite3Command(t.mainDbPath, fmt.Sprintf(sqliteBackupCommand, t.getBackupDbPath()))
	if result.Err != nil {
		return fmt.Errorf("exec backup error: exit status %d", result.ExitCode)
	}
	if err := t.checkIntegrity(t.getBackupDbPath()); err != nil {
		return errors.New("integrity check failed")
	}
	return nil
}

func (t singleDbBackupTask) backup() error {
	if err := t.prepareDb(); err != nil {
		return fmt.Errorf("failed to prepare db, %v", err)
	}
	const maxRetry = 3
	if err := retry(t.gracefulBackup, maxRetry); err != nil {
		hwlog.RunLog.Errorf("graceful backup %s failed, %v", t.getDbName(), err)
		copyFn := func() error { return t.forciblyCopy(t.mainDbPath, t.getBackupDbPath()) }
		if err := retry(copyFn, maxRetry); err != nil {
			hwlog.RunLog.Errorf("forcibly backup %s failed, %v", t.getDbName(), err)
			return fmt.Errorf("forcibly backup failed, %v", err)
		}
	}
	hwlog.RunLog.Infof("backup %s successful", t.getDbName())
	return nil
}

const (
	backupSuffix = ".backup"
)

func (t singleDbBackupTask) getBackupDbPath() string {
	return t.mainDbPath + backupSuffix
}

func (t singleDbBackupTask) getDbName() string {
	return filepath.Base(t.mainDbPath)
}

func (t singleDbBackupTask) getBackupDbName() string {
	return filepath.Base(t.getBackupDbPath())
}

func (t singleDbBackupTask) getOppositeDbPath(origin string) string {
	if strings.HasSuffix(origin, backupSuffix) {
		return t.mainDbPath
	}
	return t.getBackupDbPath()
}

func (t singleDbBackupTask) forciblyCopy(src, dst string) error {
	if err := fileutils.CopyFile(src, dst); err != nil {
		return err
	}
	if err := t.checkIntegrity(dst); err != nil {
		return err
	}
	hwlog.RunLog.Infof("forcibly copy %s successful", filepath.Base(src))
	return nil
}

func (t singleDbBackupTask) prepareDb() error {
	dbFiles := []string{t.mainDbPath, t.getBackupDbPath()}
	for _, filePath := range dbFiles {
		if err := t.prepareSingleDb(filePath); err != nil {
			return err
		}
	}
	return nil
}

var (
	regexpSqliteDbPath = regexp.MustCompile(`[^'"\\]+`)
)

func (t singleDbBackupTask) executeSqlite3Command(dbPath, command string) envutils.CommandResult {
	if !regexpSqliteDbPath.MatchString(dbPath) {
		return envutils.CommandResult{Err: errors.New("db path contains illegal chars")}
	}

	command = fmt.Sprintf("%s\n.exit\n", command)
	cmdOpts := envutils.CommandOptions{
		RunAsUser:       t.uid,
		RunAsGroup:      t.gid,
		SwitchUser:      true,
		Stdin:           []byte(command),
		WaitTimeSeconds: commandWaitTimeSeconds,
	}
	return envutils.RunCommandWithOptions(&cmdOpts, commandSqlite3, dbPath)
}

// prepareSingleDb creates an empty db and changes its permission and owner to prevent information leaking
func (t singleDbBackupTask) prepareSingleDb(filePath string) error {
	dbFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, constants.Mode600)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("failed to create emtpy db, %v", err)
		}
		linkChecker := fileutils.NewFileLinkChecker(false)
		linkChecker.SetNext(fileutils.NewFileOwnerChecker(false, false, t.uid, t.gid))
		return fileutils.SetPathPermission(filePath, constants.Mode600, false, true, linkChecker)
	}
	defer func() {
		if err := dbFile.Close(); err != nil {
			hwlog.RunLog.Errorf("failed to close db file, %v", err)
		}
	}()

	if _, err := fileutils.CheckOriginPath(filePath); err != nil {
		return fmt.Errorf("failed to check db path, %v", err)
	}
	stat, err := dbFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat db, %v", err)
	}
	if !fileutils.CheckMode(stat.Mode().Perm(), constants.ModeUmask177) {
		return errors.New("failed to check permission of db")
	}
	if err := dbFile.Chown(int(t.uid), int(t.gid)); err != nil {
		return fmt.Errorf("failed to change owner of db, %v", err)
	}
	return nil
}

func retry(fn func() error, maxRetry int) error {
	var err error
	for i := 0; i < maxRetry; i++ {
		err = fn()
		if err == nil {
			break
		}
	}
	return err
}
