// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
