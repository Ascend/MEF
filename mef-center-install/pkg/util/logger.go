// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

package util

import (
	"context"
	"fmt"
	"path"

	"huawei.com/mindx/common/hwlog"
)

func newLogConfig(LogFileName string, logBackupDir string) *hwlog.LogConfig {
	return &hwlog.LogConfig{
		LogFileName:   LogFileName,
		OnlyToFile:    true,
		MaxBackups:    hwlog.DefaultMaxBackups,
		MaxAge:        hwlog.DefaultMinSaveAge,
		IsCompress:    true,
		BackupDirName: logBackupDir,
		EscapeHtml:    true,
	}
}

// InitLogPath initialize logger
func InitLogPath(logPath string, logBackupPath string) error {
	runLogConf := newLogConfig(path.Join(logPath, RunLogFile), logBackupPath)
	opLogConf := newLogConfig(path.Join(logPath, OperateLogFile), logBackupPath)

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
