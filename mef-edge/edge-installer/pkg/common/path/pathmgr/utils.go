// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package pathmgr utils for path manager
package pathmgr

import (
	"os"
	"path/filepath"

	"huawei.com/mindx/common/fileutils"
)

// GetTargetInstallDir get the installation location of the software
func GetTargetInstallDir(installRootDir string) (string, error) {
	workPathMgr := NewWorkPathMgr(installRootDir)
	workADir := workPathMgr.GetWorkADir()
	workBDir := workPathMgr.GetWorkBDir()
	workDir := workPathMgr.GetWorkDir()

	if !fileutils.IsExist(workDir) {
		return workADir, nil
	}

	pathInfo, err := os.Lstat(workDir)
	if err != nil {
		return "", err
	}

	symlinkDir, err := filepath.EvalSymlinks(workDir)
	if err != nil {
		return "", err
	}

	if pathInfo.Mode()&os.ModeSymlink != 0 && symlinkDir == workADir && fileutils.IsExist(workADir) {
		return workBDir, nil
	}
	return workADir, nil
}
