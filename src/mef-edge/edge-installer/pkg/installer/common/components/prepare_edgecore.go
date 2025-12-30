// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package components this file for prepare edge core
package components

import (
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
)

// PrepareEdgeCore for prepare edge core
type PrepareEdgeCore struct {
	PrepareCompBase
}

// NewPrepareEdgeCore create prepare edge core instance
func NewPrepareEdgeCore(pathMgr *pathmgr.PathManager, workAbsPathMgr *pathmgr.WorkAbsPathMgr) *PrepareEdgeCore {
	return &PrepareEdgeCore{
		PrepareCompBase: PrepareCompBase{
			CompName:       constants.EdgeCore,
			PathManager:    pathMgr,
			WorkAbsPathMgr: workAbsPathMgr,
		},
	}
}

// PrepareCfgDir prepare edge core config dir
func (pec *PrepareEdgeCore) PrepareCfgDir() error {
	configDstDir := pec.SoftwarePathMgr.ConfigPathMgr.GetConfigDir()
	createDirNames := []string{constants.InnerCertPathName}
	return pec.prepareConfigDir(configDstDir, createDirNames...)
}

// Run prepare edge core
func (pec *PrepareEdgeCore) Run() error {
	var preFunc = []func() error{
		pec.prepareSoftwareDir,
		pec.prepareConfigLink,
		pec.prepareLogDirs,
		pec.prepareLogLinks,
		pec.prepareDefaultCfgBackupDir,
		pec.setOwnerAndMode,
	}
	for _, function := range preFunc {
		if err := function(); err != nil {
			hwlog.RunLog.Error(err)
			return err
		}
	}
	return nil
}
