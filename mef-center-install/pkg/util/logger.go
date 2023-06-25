// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

package util

import (
	"context"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/hwlog"
)

const (
	logMaxAge        = 30
	fileMaxSize      = 100
	opLogMaxBackups  = 10
	runLogMaxBackups = 30
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
