// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

package util

import (
	"context"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
)

const (
	logMaxAge        = 30
	fileMaxSize      = 100
	opLogMaxBackups  = 10
	runLogMaxBackups = 30

	// LogDumpRootDir is the dir to store log-dumping temp files
	LogDumpRootDir = "/var/mef_logcollect"
)

func newLogConfig(LogFileName string, logBackupDir string, maxBackups int) *hwlog.LogConfig {
	return &hwlog.LogConfig{
		LogFileName:   LogFileName,
		OnlyToFile:    true,
		MaxBackups:    maxBackups,
		MaxAge:        logMaxAge,
		IsCompress:    true,
		BackupDirName: logBackupDir,
		EscapeHtml:    true,
		FileMaxSize:   fileMaxSize,
	}
}

// InitLogPath initialize logger
func InitLogPath(logPath string, logBackupPath string) error {
	runLogConf := newLogConfig(filepath.Join(logPath, RunLogFile), logBackupPath, runLogMaxBackups)
	runLogConf.DisableRotationIfSwitchUser = true
	opLogConf := newLogConfig(filepath.Join(logPath, OperateLogFile), logBackupPath, opLogMaxBackups)

	if err := initHwLogger(runLogConf, opLogConf); err != nil {
		return fmt.Errorf("initialize hwlog failed, error: %v", err.Error())
	}

	return nil
}

func initHwLogger(runLogConfig, opLogConfig *hwlog.LogConfig) error {
	if err := hwlog.InitRunLogger(runLogConfig, context.Background()); err != nil {
		return err
	}
	if err := hwlog.InitOperateLogger(opLogConfig, context.Background()); err != nil {
		return err
	}
	return nil
}

// PrepareLogDumpDir prepares the log dump dir
func PrepareLogDumpDir() error {
	if _, err := fileutils.RealDirCheck(filepath.Dir(LogDumpRootDir), true, false); err != nil {
		return fmt.Errorf("check parent dir of log dump dir failed, %v", err)
	}

	if err := fileutils.DeleteAllFileWithConfusion(LogDumpRootDir); err != nil {
		return fmt.Errorf("delete log dump dir failed, %v", err)
	}
	if err := fileutils.CreateDir(LogDumpRootDir, common.Mode700); err != nil {
		return fmt.Errorf("create log dump dir failed, %v", err)
	}
	uid, gid, err := GetMefId()
	if err != nil {
		return fmt.Errorf("get mef id failed, %v", err)
	}
	if err := fileutils.SetPathOwnerGroup(
		fileutils.SetOwnerParam{Path: LogDumpRootDir, Uid: uid, Gid: gid, IgnoreFile: true}); err != nil {
		return fmt.Errorf("set log dump dir ownership failed, %v", err)
	}
	return nil
}
