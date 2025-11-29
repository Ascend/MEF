// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_A500

// Package components for prepare default config backup
package components

import (
	"fmt"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
)

func (pi *PrepareInstaller) prepareDefaultCfgBackupDir() error {
	return pi.prepareDefaultCfgBackupDirBase()
}

func (peo *PrepareEdgeOm) prepareDefaultCfgBackupDir() error {
	createDirNames := []string{constants.InnerCertPathName, constants.ImageCertPathName}
	return peo.prepareDefaultCfgBackupDirBase(createDirNames...)
}

func (pem *PrepareEdgeMain) prepareDefaultCfgBackupDir() error {
	createDirNames := []string{constants.InnerCertPathName, constants.PeerCerts}
	return pem.prepareDefaultCfgBackupDirBase(createDirNames...)
}

func (pec *PrepareEdgeCore) prepareDefaultCfgBackupDir() error {
	createDirNames := []string{constants.InnerCertPathName}
	return pec.prepareDefaultCfgBackupDirBase(createDirNames...)
}

func (pcb *PrepareCompBase) prepareDefaultCfgBackupDirBase(createDirNames ...string) error {
	cfgBackupDir := pcb.SoftwarePathMgr.ConfigPathMgr.GetConfigBackupDir()
	cfgBackupTempDir := pcb.SoftwarePathMgr.ConfigPathMgr.GetConfigBackupTempDir()
	if fileutils.IsExist(cfgBackupTempDir) {
		cfgBackupDir = cfgBackupTempDir
	}
	if err := fileutils.CreateDir(cfgBackupDir, constants.Mode755); err != nil {
		return fmt.Errorf("create dir [%s] failed, error: %v", cfgBackupDir, err)
	}

	if err := pcb.prepareConfigDir(cfgBackupDir, createDirNames...); err != nil {
		hwlog.RunLog.Errorf("prepare %s config backup dir failed, error: %v", pcb.CompName, err)
		return fmt.Errorf("prepare %s config backup dir failed", pcb.CompName)
	}
	hwlog.RunLog.Infof("prepare %s config backup dir success", pcb.CompName)
	return nil
}
