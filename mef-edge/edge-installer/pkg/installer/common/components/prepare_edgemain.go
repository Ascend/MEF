// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package components this file for prepare edge main
package components

import (
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
)

// PrepareEdgeMain for prepare edge main
type PrepareEdgeMain struct {
	PrepareCompBase
}

// NewPrepareEdgeMain create prepare edge main instance
func NewPrepareEdgeMain(pathMgr *pathmgr.PathManager, workAbsPathMgr *pathmgr.WorkAbsPathMgr) *PrepareEdgeMain {
	return &PrepareEdgeMain{
		PrepareCompBase: PrepareCompBase{
			CompName:       constants.EdgeMain,
			PathManager:    pathMgr,
			WorkAbsPathMgr: workAbsPathMgr,
		},
	}
}

// PrepareCfgDir prepare edge main config dir
func (pem *PrepareEdgeMain) PrepareCfgDir() error {
	configDstDir := pem.SoftwarePathMgr.ConfigPathMgr.GetConfigDir()
	createDirNames := []string{constants.InnerCertPathName, constants.PeerCerts}
	return pem.prepareConfigDir(configDstDir, createDirNames...)
}

// Run prepare edge main
func (pem *PrepareEdgeMain) Run() error {
	var preFunc = []func() error{
		pem.prepareSoftwareDir,
		pem.prepareConfigLink,
		pem.prepareLogDirs,
		pem.prepareLogLinks,
		pem.prepareDefaultCfgBackupDir,
		pem.setOwnerAndMode,
	}
	for _, function := range preFunc {
		if err := function(); err != nil {
			hwlog.RunLog.Error(err)
			return err
		}
	}
	return nil
}
