// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import "path"

// LogDirPathMgr is a struct that controls all dir/file path in the log dir
// all dir/file path in the log dir should be got by specified func in this struct
type LogDirPathMgr struct {
	rootPath string
}

// GetLogRootPath returns the log root path
func (ldm *LogDirPathMgr) GetLogRootPath() string {
	return ldm.rootPath
}

// GetModuleLogPath returns the mef-center_log dir path
func (ldm *LogDirPathMgr) GetModuleLogPath() string {
	return path.Join(ldm.rootPath, ModuleLogName)
}

// GetInstallLogPath returns the installation dir path
func (ldm *LogDirPathMgr) GetInstallLogPath() string {
	return path.Join(ldm.GetModuleLogPath(), InstallLogDir)
}

// GetComponentLogPath returns a single component's log dir path by component's name
func (ldm *LogDirPathMgr) GetComponentLogPath(component string) string {
	return path.Join(ldm.GetModuleLogPath(), component)
}

// InitLogDirPathMgr returns the LogDirPathMgr construct by the log root path
// Each call to this func returns a distinct mgr value even if the log path is identical
func InitLogDirPathMgr(rootPath string) *LogDirPathMgr {
	return &LogDirPathMgr{rootPath: rootPath}
}
