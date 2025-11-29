// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

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
