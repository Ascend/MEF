// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package nginxlogrotate enables nginx to rotate and backup logs.
package nginxlogrotate

import (
	"errors"
	"os"
	"path/filepath"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/logmgmt/logrotate"
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
	if err := fileutils.CreateDir(defaultLogBackupDir, defaultPermission); err != nil && errors.Is(err, os.ErrExist) {
		return logrotate.Configs{}, err
	}
	if err := fileutils.SetPathPermission(defaultLogBackupDir, defaultPermission, false, false); err != nil {
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
