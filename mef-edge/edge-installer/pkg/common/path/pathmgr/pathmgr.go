// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
