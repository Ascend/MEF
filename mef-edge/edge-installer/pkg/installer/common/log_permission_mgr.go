// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package common for log path permissions manager
package common

import (
	"fmt"
	"os"
	"path/filepath"

	"huawei.com/mindx/common/fileutils"

	"edge-installer/pkg/common/constants"
)

// LogPermissionMgr permission manager for setting owner,group and modeUmask for log directories
type LogPermissionMgr struct {
	LogPath string
}

// CheckPermission check log path permissions
func (lpm LogPermissionMgr) CheckPermission() error {
	if !fileutils.IsExist(lpm.LogPath) {
		return nil
	}
	logPermList := lpm.GetLogPermList()
	for _, logPerm := range logPermList {
		if !fileutils.IsExist(logPerm.dir) {
			continue
		}
		if err := lpm.checkPathPermission(logPerm.dir, logPerm.modeUmask, logPerm.userUid); err != nil {
			return err
		}
		if err := lpm.checkFilePermission(logPerm.dir, constants.ModeUmask137, logPerm.userUid); err != nil {
			return err
		}
	}
	if err := lpm.checkEdgeMainPermission(); err != nil {
		return err
	}
	return nil
}

func (lpm LogPermissionMgr) checkEdgeMainPermission() error {
	edgeMainDir := filepath.Join(lpm.LogPath, constants.EdgeMain)
	if !fileutils.IsExist(edgeMainDir) {
		return nil
	}
	edgeMainPerm, err := lpm.getEdgeMainPerm()
	if err != nil {
		return fmt.Errorf("check path [%s] owner and permission failed: %v", edgeMainDir, err)
	}
	if err = lpm.checkPathPermission(edgeMainPerm.dir, edgeMainPerm.modeUmask, edgeMainPerm.userUid); err != nil {
		return err
	}
	if err = lpm.checkFilePermission(edgeMainPerm.dir, constants.ModeUmask137, edgeMainPerm.userUid); err != nil {
		return err
	}
	return nil
}

func (lpm LogPermissionMgr) checkFilePermission(dir string, mode os.FileMode, uid uint32) error {
	checkLogFiles := []string{
		fmt.Sprintf("%s_%s", filepath.Base(dir), constants.RunLogFile),
		fmt.Sprintf("%s_%s", filepath.Base(dir), constants.OperateLogFile),
		constants.ResetLogFile,
	}
	for _, logFile := range checkLogFiles {
		logFilePath := filepath.Join(dir, logFile)
		if !fileutils.IsExist(logFilePath) {
			continue
		}
		if err := lpm.checkPathPermission(logFilePath, mode, uid); err != nil {
			return err
		}
	}
	return nil
}

func (lpm LogPermissionMgr) checkPathPermission(path string, mode os.FileMode, uid uint32) error {
	if _, err := fileutils.CheckOriginPath(path); err != nil {
		return fmt.Errorf("check path [%s] failed, error: %v", path, err)
	}

	if _, err := fileutils.CheckOwnerAndPermission(path, mode, uid); err != nil {
		return fmt.Errorf("check path [%s] owner and permission failed, error: %v", path, err)
	}
	return nil
}
