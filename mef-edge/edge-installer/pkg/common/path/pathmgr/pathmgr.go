// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package pathmgr path manager
package pathmgr

// PathManager path manager, including software path manager and log path manager
type PathManager struct {
	*SoftwarePathMgr
	*LogPathMgr
}

// NewPathMgr new path manager
func NewPathMgr(installRootDir, installationPkgDir, logRootDir, logBackupRootDir string) *PathManager {
	return &PathManager{
		SoftwarePathMgr: NewSoftwarePathMgr(installRootDir, installationPkgDir),
		LogPathMgr:      NewLogPathMgr(logRootDir, logBackupRootDir),
	}
}
