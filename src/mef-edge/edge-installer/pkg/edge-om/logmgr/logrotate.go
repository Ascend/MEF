// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package logmgr
package logmgr

import (
	"huawei.com/mindx/common/logmgmt/logrotate"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
)

const (
	defaultCheckInterval             = 60
	defaultEdgeCoreLogMaxBackups     = 30
	defaultEdgeCoreLogMaxSize        = 20
	defaultDevicePluginLogMaxBackups = 30
	defaultDevicePluginMaxSize       = 12
)

func newLogRotator(logPathMgr *pathmgr.LogPathMgr) *logrotate.LogRotator {
	edgeCoreLog := logrotate.Config{
		LogFile:    logPathMgr.GetComponentLogPath(constants.EdgeCore, constants.EdgeCoreLogFile),
		BackupDir:  logPathMgr.GetComponentLogBackupDir(constants.EdgeCore),
		MaxBackups: defaultEdgeCoreLogMaxBackups,
		MaxSizeMB:  defaultEdgeCoreLogMaxSize,
	}
	devicePluginLog := logrotate.Config{
		LogFile:    logPathMgr.GetComponentLogPath(constants.DevicePlugin, constants.DevicePluginLogFile),
		BackupDir:  logPathMgr.GetComponentLogBackupDir(constants.DevicePlugin),
		MaxBackups: defaultDevicePluginLogMaxBackups,
		MaxSizeMB:  defaultDevicePluginMaxSize,
	}
	configs := logrotate.Configs{
		CheckIntervalSeconds: defaultCheckInterval,
		Logs:                 []logrotate.Config{edgeCoreLog, devicePluginLog},
	}
	return logrotate.New(configs)
}
