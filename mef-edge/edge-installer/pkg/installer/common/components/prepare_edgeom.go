// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package components this file for prepare edge om
package components

import (
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
)

// PrepareEdgeOm for prepare edge_om
type PrepareEdgeOm struct {
	PrepareCompBase
}

// NewPrepareEdgeOm create prepare edge om instance
func NewPrepareEdgeOm(pathMgr *pathmgr.PathManager, workAbsPathMgr *pathmgr.WorkAbsPathMgr) *PrepareEdgeOm {
	return &PrepareEdgeOm{
		PrepareCompBase: PrepareCompBase{
			CompName:       constants.EdgeOm,
			PathManager:    pathMgr,
			WorkAbsPathMgr: workAbsPathMgr,
		},
	}
}

// PrepareCfgDir prepare edge om config dir
func (peo *PrepareEdgeOm) PrepareCfgDir() error {
	configDstDir := peo.SoftwarePathMgr.ConfigPathMgr.GetConfigDir()
	createDirNames := []string{constants.InnerCertPathName, constants.ImageCertPathName}
	return peo.prepareConfigDir(configDstDir, createDirNames...)
}

// Run prepare edge om
func (peo *PrepareEdgeOm) Run() error {
	var preFunc = []func() error{
		peo.prepareSoftwareDir,
		peo.prepareConfigLink,
		peo.prepareLogDirs,
		peo.prepareLogLinks,
		peo.prepareDefaultCfgBackupDir,
		peo.setOwnerAndMode,
	}
	for _, function := range preFunc {
		if err := function(); err != nil {
			hwlog.RunLog.Error(err)
			return err
		}
	}
	return nil
}
