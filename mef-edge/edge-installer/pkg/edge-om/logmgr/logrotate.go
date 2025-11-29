// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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
