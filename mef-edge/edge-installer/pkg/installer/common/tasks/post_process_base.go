// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package tasks for post process base
package tasks

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/common"
)

// PostProcessBaseTask the task for post process base task
type PostProcessBaseTask struct {
	WorkPathMgr *pathmgr.WorkPathMgr
	LogPathMgr  *pathmgr.LogPathMgr
}

// RemoveUpgradeBinByPath Remove upgrade binary file
func (p *PostProcessBaseTask) RemoveUpgradeBinByPath(upgradeBin string) error {
	if err := fileutils.DeleteFile(upgradeBin); err != nil {
		return fmt.Errorf("remove upgrade binary file [%s] failed, error: %v", upgradeBin, err)
	}

	hwlog.RunLog.Info("remove upgrade binary file success")
	return nil
}

// CreateSoftwareSymlink create software symlink
func (p *PostProcessBaseTask) CreateSoftwareSymlink() error {
	softwareDir, err := pathmgr.GetTargetInstallDir(p.WorkPathMgr.GetInstallRootDir())
	if err != nil {
		return fmt.Errorf("get target software dir failed, error: %v", err)
	}

	if fileutils.IsExist(p.WorkPathMgr.GetUpgradeTempDir()) {
		if err = p.removeBackupDir(softwareDir); err != nil {
			return err
		}
		if err = p.renameUpgradeDir(softwareDir); err != nil {
			return err
		}
	}

	softwareDirSymlink := p.WorkPathMgr.GetWorkDir()
	if err = fileutils.DeleteFile(softwareDirSymlink); err != nil {
		return fmt.Errorf("remove old software dir symlink failed, error: %v", err)
	}

	if err = os.Symlink(softwareDir, softwareDirSymlink); err != nil {
		return fmt.Errorf("create new software dir symlink failed, error: %v", err)
	}

	hwlog.RunLog.Info("create software dir symlink success")
	return nil
}

// UpdateMefServiceInfo update service info
func (p *PostProcessBaseTask) UpdateMefServiceInfo() error {
	mgr := common.NewComponentMgr(p.WorkPathMgr.GetInstallRootDir())
	if err := mgr.UpdateServiceFiles(p.LogPathMgr.GetEdgeLogDir(), p.LogPathMgr.GetEdgeLogBackupDir()); err != nil {
		return fmt.Errorf("update service files failed, error: %v", err)
	}

	if err := mgr.RegisterAllServices(); err != nil {
		return fmt.Errorf("register all services failed, error: %v", err)
	}

	hwlog.RunLog.Info("update mef service info success")
	return nil
}

// SetSoftwareDirImmutable set software dir immutable
func (p *PostProcessBaseTask) SetSoftwareDirImmutable() error {
	workRealPath, err := filepath.EvalSymlinks(p.WorkPathMgr.GetWorkDir())
	if err != nil {
		hwlog.RunLog.Errorf("get software real path failed, error: %v", err)
		return err
	}

	if err = util.SetImmutable(workRealPath); err != nil {
		hwlog.RunLog.Warnf("set software path [%s] immutable find errors, maybe include link files", workRealPath)
	}

	components := []string{constants.EdgeInstaller, constants.EdgeOm, constants.EdgeMain, constants.EdgeCore,
		constants.DevicePlugin}
	for _, comp := range components {
		compVarPath := pathmgr.NewWorkAbsPathMgr(workRealPath).GetCompVarDir(comp)
		if err = util.UnSetImmutable(compVarPath); err != nil {
			hwlog.RunLog.Warnf("unset no need immutable path [%s] failed, maybe include link files", compVarPath)
		}
	}

	hwlog.RunLog.Info("set software immutable success")
	return nil
}

func (p *PostProcessBaseTask) removeBackupDir(backupDir string) error {
	if err := util.UnSetImmutable(backupDir); err != nil {
		hwlog.RunLog.Warnf("unset backup dir[%s] immutable find errors, maybe include link files", backupDir)
	}
	if err := fileutils.DeleteAllFileWithConfusion(backupDir); err != nil {
		hwlog.RunLog.Errorf("remove backup directory failed, error: %v", err)
		return errors.New("remove backup directory failed")
	}
	return nil
}

func (p *PostProcessBaseTask) renameUpgradeDir(renameTo string) error {
	upgradeTempDir := p.WorkPathMgr.GetUpgradeTempDir()
	if err := fileutils.RenameFile(upgradeTempDir, renameTo); err != nil {
		hwlog.RunLog.Errorf("rename upgrade temp directory name to [%s] failed, error: %v", renameTo, err)
		return fmt.Errorf("rename upgrade temp directory name to [%s] failed", renameTo)
	}
	return nil
}
