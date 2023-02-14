// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import "path"

// LogDirPathMgr is a struct that controls all dir/file path in the log dir
// all dir/file path in the log dir should be got by specified func in this struct
type LogDirPathMgr struct {
	logRootPath       string
	logBackupRootPath string
}

// GetLogRootPath returns the log root path
func (ldm *LogDirPathMgr) GetLogRootPath() string {
	return ldm.logRootPath
}

// GetModuleLogPath returns the mef-center-log dir path
func (ldm *LogDirPathMgr) GetModuleLogPath() string {
	return path.Join(ldm.logRootPath, ModuleLogName)
}

// GetInstallLogPath returns the installation dir path
func (ldm *LogDirPathMgr) GetInstallLogPath() string {
	return path.Join(ldm.GetModuleLogPath(), InstallLogDir)
}

// GetComponentLogPath returns a single component's log dir path by component's name
func (ldm *LogDirPathMgr) GetComponentLogPath(component string) string {
	return path.Join(ldm.GetModuleLogPath(), component)
}

// GetLogBackupRootPath returns the root path of log backup files
func (ldm *LogDirPathMgr) GetLogBackupRootPath() string {
	return ldm.logBackupRootPath
}

// GetModuleLogBackupPath returns the mef-center-log dir path
func (ldm *LogDirPathMgr) GetModuleLogBackupPath() string {
	return path.Join(ldm.logBackupRootPath, ModuleLogBackupName)
}

// GetInstallLogBackupPath returns the installation dir path
func (ldm *LogDirPathMgr) GetInstallLogBackupPath() string {
	return path.Join(ldm.GetModuleLogBackupPath(), InstallLogDir)
}

// GetComponentBackupLogPath returns a single component's path of log backup files by component's name
func (ldm *LogDirPathMgr) GetComponentBackupLogPath(component string) string {
	return path.Join(ldm.GetModuleLogBackupPath(), component)
}

// InitLogDirPathMgr returns the LogDirPathMgr construct by the dir of log backup files
// Each call to this func returns a distinct mgr value even if the backup path is identical
func InitLogDirPathMgr(rootPath string, logBackupRootPath string) *LogDirPathMgr {
	return &LogDirPathMgr{logRootPath: rootPath, logBackupRootPath: logBackupRootPath}
}
