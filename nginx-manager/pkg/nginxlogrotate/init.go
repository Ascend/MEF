// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nginxlogrotate enables nginx to rotate and backup logs.
package nginxlogrotate

import (
	"os"
	"path/filepath"

	"huawei.com/mindxedge/base/common/logmgmt/logrotate"
)

const (
	defaultCheckInterval       = 60
	defaultLogDir              = "/home/MEFCenter/logs/"
	defaultLogBackupDir        = "/home/MEFCenter/logs_backup/"
	defaultAccessLogFileName   = "access.log"
	defaultAccessLogMaxBackups = 30
	defaultAccessLogMaxSize    = 90
	defaultErrorLogFileName    = "error.log"
	defaultErrorLogMaxBackups  = 30
	defaultErrorLogMaxSize     = 10
	defaultPermission          = 0750
)

// Setup evaluates log rotate configurations and does basic preparation.
func Setup() (logrotate.Configs, error) {
	if err := os.MkdirAll(defaultLogBackupDir, defaultPermission); err != nil && os.IsExist(err) {
		return logrotate.Configs{}, err
	}
	if err := os.Chmod(defaultLogBackupDir, defaultPermission); err != nil {
		return logrotate.Configs{}, err
	}
	accessLog := logrotate.Config{
		LogFile:    filepath.Join(defaultLogDir, defaultAccessLogFileName),
		BackupDir:  defaultLogBackupDir,
		MaxBackups: defaultAccessLogMaxBackups,
		MaxSizeMB:  defaultAccessLogMaxSize,
	}
	errorLog := logrotate.Config{
		LogFile:    filepath.Join(defaultLogDir, defaultErrorLogFileName),
		BackupDir:  defaultLogBackupDir,
		MaxBackups: defaultErrorLogMaxBackups,
		MaxSizeMB:  defaultErrorLogMaxSize,
	}
	return logrotate.Configs{
		CheckIntervalSeconds: defaultCheckInterval,
		Logs:                 []logrotate.Config{accessLog, errorLog},
	}, nil
}
