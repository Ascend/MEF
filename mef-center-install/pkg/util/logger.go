// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

package util

import (
	"context"
	"fmt"
	"path"
	"path/filepath"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/mef-center-install/pkg/install"
)

func newLogConfig(LogFileName string) *hwlog.LogConfig {
	return &hwlog.LogConfig{
		LogFileName: LogFileName,
		OnlyToFile:  true,
		MaxBackups:  hwlog.DefaultMaxBackups,
		MaxAge:      hwlog.DefaultMinSaveAge,
	}
}

// InitLogPath initialize logger
func InitLogPath(logPath string) error {
	realLogPath, err := filepath.EvalSymlinks(logPath)
	if err != nil {
		return fmt.Errorf("get the real path of log path [%s] failed, error: %s", logPath, err.Error())
	}

	runLogConf := newLogConfig(path.Join(realLogPath, install.RunLogFile))
	opLogConf := newLogConfig(path.Join(realLogPath, install.OperateLogFile))

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
