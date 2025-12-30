// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package components this file for prepare components base
package components

import (
	"fmt"
	"os"
	"path/filepath"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/installer/common"
)

// PrepareCompBase prepare components base
type PrepareCompBase struct {
	CompName string
	*pathmgr.PathManager
	*pathmgr.WorkAbsPathMgr
}

func (pcb *PrepareCompBase) prepareConfigDir(configDstDir string, createDirNames ...string) error {
	configSrc := pcb.SoftwarePathMgr.GetPkgCompConfigDir(pcb.CompName)
	configDst := filepath.Join(configDstDir, pcb.CompName)
	if err := fileutils.CopyDir(configSrc, configDst); err != nil {
		return fmt.Errorf("copy %s config dir failed, error: %v", pcb.CompName, err)
	}

	for _, createDirName := range createDirNames {
		createDir := filepath.Join(configDst, createDirName)
		if err := fileutils.CreateDir(createDir, constants.Mode700); err != nil {
			return fmt.Errorf("create dir [%s] failed, error: %v", createDir, err)
		}
	}
	return nil
}

func (pcb *PrepareCompBase) prepareSoftwareDir() error {
	installerSrc := pcb.SoftwarePathMgr.GetPkgCompSoftwareDir(pcb.CompName)
	installerDst := pcb.WorkAbsPathMgr.GetCompWorkDir(pcb.CompName)
	if err := fileutils.CopyDir(installerSrc, installerDst); err != nil {
		return fmt.Errorf("copy %s software dir failed, error: %v", pcb.CompName, err)
	}

	hwlog.RunLog.Infof("prepare %s software dir success", pcb.CompName)
	return nil
}

func (pcb *PrepareCompBase) prepareLogDirs() error {
	compLogDirs := []string{
		pcb.LogPathMgr.GetComponentLogDir(pcb.CompName),
		pcb.LogPathMgr.GetComponentLogBackupDir(pcb.CompName),
	}
	for _, compLogDir := range compLogDirs {
		if err := fileutils.CreateDir(compLogDir, constants.Mode750); err != nil {
			return fmt.Errorf("create log dir %s failed, error: %v", compLogDir, err)
		}
	}

	hwlog.RunLog.Infof("create %s log dirs success", pcb.CompName)
	return nil
}

func (pcb *PrepareCompBase) prepareLogLinks() error {
	varDir := pcb.WorkAbsPathMgr.GetCompVarDir(pcb.CompName)
	if err := fileutils.CreateDir(varDir, constants.Mode700); err != nil {
		return fmt.Errorf("create %s var dir failed, error: %v", pcb.CompName, err)
	}

	symlinkList := []struct {
		src string
		dst string
	}{
		{src: pcb.LogPathMgr.GetComponentLogDir(pcb.CompName), dst: filepath.Join(varDir, constants.Log)},
		{src: pcb.LogPathMgr.GetComponentLogBackupDir(pcb.CompName), dst: filepath.Join(varDir, constants.LogBackup)},
	}
	for _, symlinkPair := range symlinkList {
		if err := os.Symlink(symlinkPair.src, symlinkPair.dst); err != nil {
			return fmt.Errorf("create log dir symlink %s failed, error: %v", symlinkPair.dst, err)
		}
	}

	hwlog.RunLog.Infof("create %s log dir symlinks success", pcb.CompName)
	return nil
}

func (pcb *PrepareCompBase) prepareConfigLink() error {
	configSrc := pcb.SoftwarePathMgr.ConfigPathMgr.GetCompConfigDir(pcb.CompName)
	configDst := pcb.WorkAbsPathMgr.GetCompConfigDir(pcb.CompName)
	if err := os.Symlink(configSrc, configDst); err != nil {
		return fmt.Errorf("create %s config dir symlink failed, error: %v", pcb.CompName, err)
	}

	hwlog.RunLog.Infof("create %s config dir symlink success", pcb.CompName)
	return nil
}

func (pcb *PrepareCompBase) setOwnerAndMode() error {
	workAbsDir, err := pcb.SoftwarePathMgr.WorkPathMgr.GetWorkAbsDir()
	if err != nil {
		return err
	}

	permMgr := common.PermissionMgr{
		CompName:       pcb.CompName,
		ConfigPathMgr:  pcb.ConfigPathMgr,
		WorkAbsPathMgr: pathmgr.NewWorkAbsPathMgr(workAbsDir),
		LogPathMgr:     pcb.LogPathMgr,
	}
	if err := permMgr.SetOwnerAndGroup(); err != nil {
		return fmt.Errorf("set %s owner failed", pcb.CompName)
	}

	if err := permMgr.SetMode(); err != nil {
		return fmt.Errorf("set %s mode failed", pcb.CompName)
	}

	hwlog.RunLog.Infof("set %s owner and mode success", pcb.CompName)
	return nil
}
