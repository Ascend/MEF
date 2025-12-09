// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package components this file for prepare edge installer
package components

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
)

// PrepareInstaller for prepare edge installer
type PrepareInstaller struct {
	PrepareCompBase
}

// NewPrepareInstaller create prepare installer instance
func NewPrepareInstaller(pathMgr *pathmgr.PathManager, workAbsPathMgr *pathmgr.WorkAbsPathMgr) *PrepareInstaller {
	return &PrepareInstaller{
		PrepareCompBase: PrepareCompBase{
			CompName:       constants.EdgeInstaller,
			PathManager:    pathMgr,
			WorkAbsPathMgr: workAbsPathMgr,
		},
	}
}

// PrepareCfgDir prepare edge installer config dir
func (pi *PrepareInstaller) PrepareCfgDir() error {
	return pi.prepareConfigDir(pi.SoftwarePathMgr.ConfigPathMgr.GetConfigDir())
}

// Run prepare installer
func (pi *PrepareInstaller) Run() error {
	var preFunc = []func() error{
		pi.prepareVersionFile,
		pi.prepareSoftwareDir,
		pi.prepareRunSh,
		pi.prepareLib,
		pi.removeUnnecessaryFiles,
		pi.prepareConfigLink,
		pi.prepareLogDirs,
		pi.prepareLogLinks,
		pi.prepareDefaultCfgBackupDir,
		pi.setOwnerAndMode,
	}
	for _, function := range preFunc {
		if err := function(); err != nil {
			hwlog.RunLog.Error(err)
			return err
		}
	}
	return nil
}

func (pi *PrepareInstaller) prepareVersionFile() error {
	versionSrc := pi.SoftwarePathMgr.GetPkgVersionXmlPath()
	versionDst := pi.WorkAbsPathMgr.GetVersionXmlPath()

	if err := fileutils.CopyFile(versionSrc, versionDst); err != nil {
		return fmt.Errorf("copy %s failed, error: %v", constants.VersionXml, err)
	}

	if err := fileutils.SetPathPermission(versionDst, constants.Mode400, false, false); err != nil {
		return fmt.Errorf("set file [%s] mode failed: %v", versionDst, err)
	}

	hwlog.RunLog.Infof("prepare %s file success", constants.VersionXml)
	return nil
}

// removeUnnecessaryFiles install, upgrade, and factory restoration files will no longer be used after installation
func (pi *PrepareInstaller) removeUnnecessaryFiles() error {
	files := []string{
		pi.WorkAbsPathMgr.GetInstallBinaryPath(),
		pi.WorkAbsPathMgr.GetUpgradeShPath(),
		pi.WorkAbsPathMgr.GetResetInstallShPath(),
		pi.WorkAbsPathMgr.GetServicePath(constants.ResetService),
	}
	for _, f := range files {
		if err := fileutils.DeleteFile(f); err != nil {
			return fmt.Errorf("remove file [%s] failed, error: %v", f, err)
		}
	}

	hwlog.RunLog.Info("remove unnecessary files success")
	return nil
}

func (pi *PrepareInstaller) prepareRunSh() error {
	runSrc := pi.SoftwarePathMgr.GetPkgRunShPath()
	runDst := pi.WorkAbsPathMgr.GetRunShPath()
	if err := fileutils.CopyFile(runSrc, runDst); err != nil {
		return fmt.Errorf("copy file [%s] failed, error: %v", constants.RunScript, err)
	}
	if err := fileutils.SetPathPermission(runDst, constants.Mode500, false, false); err != nil {
		return fmt.Errorf("set file [%s] mode failed: %v", constants.RunScript, err)
	}

	hwlog.RunLog.Info("prepare run.sh success")
	return nil
}

func (pi *PrepareInstaller) prepareLib() error {
	libSrc := pi.SoftwarePathMgr.GetPkgLibDir()
	libDst := pi.WorkAbsPathMgr.GetLibDir()

	if err := fileutils.CopyDirWithSoftlink(libSrc, libDst, &fileutils.FileBaseChecker{}); err != nil {
		hwlog.RunLog.Errorf("copy lib dir failed: %v", err)
		return errors.New("copy lib dir failed")
	}

	if err := fileutils.SetPathPermission(libDst, fileutils.Mode444, true, false,
		&fileutils.FileBaseChecker{}); err != nil {
		return fmt.Errorf("set lib files mode in dir [%s] failed: %v", libDst, err)
	}
	if err := fileutils.SetPathPermission(libDst, fileutils.Mode755, false, false); err != nil {
		return fmt.Errorf("set dir [%s] mode failed: %v", libDst, err)
	}

	hwlog.RunLog.Info("prepare lib dir success")
	return nil
}
