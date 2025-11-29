// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package tasks for setting work path
package tasks

import (
	"fmt"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
)

// SetWorkPathTask the task for prepare install work path
type SetWorkPathTask struct {
	PathMgr *pathmgr.PathManager
}

// Run set work path task
func (swp *SetWorkPathTask) Run() error {
	var setFunc = []func() error{
		swp.setRootDirsParentPerm,
		swp.prepareInstallDir,
		swp.prepareLogDirs,
	}
	for _, function := range setFunc {
		if err := function(); err != nil {
			hwlog.RunLog.Error(err)
			return err
		}
	}
	return nil
}

func (swp *SetWorkPathTask) setRootDirsParentPerm() error {
	dirs := []string{swp.PathMgr.GetInstallRootDir(), swp.PathMgr.GetLogRootDir(), swp.PathMgr.GetLogBackupRootDir()}
	for _, dir := range dirs {
		if err := fileutils.SetParentPathPermission(dir, constants.Mode755); err != nil {
			return fmt.Errorf("set dir [%s] parent permission failed, error: %v", dir, err)
		}
		hwlog.RunLog.Infof("set dir [%s] parent permission success", dir)
	}
	return nil
}

func (swp *SetWorkPathTask) prepareInstallDir() error {
	workAbsDir, err := swp.PathMgr.WorkPathMgr.GetWorkAbsDir()
	if err != nil {
		return fmt.Errorf("get software abs dir failed, error: %v", err)
	}

	dirs := []string{workAbsDir, swp.PathMgr.ConfigPathMgr.GetConfigDir()}
	for _, dir := range dirs {
		if err = fileutils.CreateDir(dir, constants.Mode755); err != nil {
			return fmt.Errorf("create dir [%s] failed, error: %v", dir, err)
		}
		hwlog.RunLog.Infof("prepare dir [%s] success", dir)
	}
	return nil
}

func (swp *SetWorkPathTask) prepareLogDirs() error {
	logDirs := []string{swp.PathMgr.GetEdgeLogDir(), swp.PathMgr.GetEdgeLogBackupDir()}
	for _, logDir := range logDirs {
		if err := fileutils.CreateDir(logDir, constants.Mode755); err != nil {
			return fmt.Errorf("create log dir [%s] failed, error: %v", logDir, err)
		}

		if err := fileutils.SetPathPermission(logDir, constants.Mode755, false, false); err != nil {
			return fmt.Errorf("set log dir [%s] permission failed, error: %v", logDir, err)
		}
		hwlog.RunLog.Infof("prepare log dir [%s] success", logDir)
	}
	return nil
}
