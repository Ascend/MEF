// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package pathmgr log path manager
package pathmgr

import (
	"path/filepath"

	"edge-installer/pkg/common/constants"
)

// LogPathMgr log path manager
type LogPathMgr struct {
	logRootDir       string
	logBackupRootDir string
}

// NewLogPathMgr new log path manager
func NewLogPathMgr(logRootDir, logBackupRootDir string) *LogPathMgr {
	return &LogPathMgr{logRootDir: logRootDir, logBackupRootDir: logBackupRootDir}
}

// GetLogRootDir get log root dir. default: /var/alog
func (ldm *LogPathMgr) GetLogRootDir() string {
	return ldm.logRootDir
}

// GetEdgeLogDir get edge log dir. default: /var/alog/MEFEdge_log
func (ldm *LogPathMgr) GetEdgeLogDir() string {
	return filepath.Join(ldm.GetLogRootDir(), constants.MEFEdgeLogName)
}

// GetComponentLogDir get component log dir. e.g. /var/alog/MEFEdge_log/edge_installer
func (ldm *LogPathMgr) GetComponentLogDir(component string) string {
	return filepath.Join(ldm.GetEdgeLogDir(), component)
}

// GetComponentLogPath get component log file path. e.g. /var/alog/MEFEdge_log/edge_installer/edge_installer_run.log
func (ldm *LogPathMgr) GetComponentLogPath(component, fileName string) string {
	return filepath.Join(ldm.GetComponentLogDir(component), fileName)
}

// GetLogBackupRootDir get log backup root dir. default: /home/log
func (ldm *LogPathMgr) GetLogBackupRootDir() string {
	return ldm.logBackupRootDir
}

// GetEdgeLogBackupDir get edge log backup dir. default: /home/log/MEFEdge_logbackup
func (ldm *LogPathMgr) GetEdgeLogBackupDir() string {
	return filepath.Join(ldm.GetLogBackupRootDir(), constants.MEFEdgeLogBackupName)
}

// GetComponentLogBackupDir get component log backup dir. e.g. /home/log/MEFEdge_logbackup/edge_installer
func (ldm *LogPathMgr) GetComponentLogBackupDir(component string) string {
	return filepath.Join(ldm.GetEdgeLogBackupDir(), component)
}

// GetComponentLogBackupPath get component log backup file path.
// e.g. /home/log/MEFEdge_logbackup/edge_installer/edge_installer_run.log
func (ldm *LogPathMgr) GetComponentLogBackupPath(component, fileName string) string {
	return filepath.Join(ldm.GetComponentLogBackupDir(component), fileName)
}
