// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package common this file for manage permission of work path when install or upgrade
package common

import (
	"errors"
	"os"
	"path/filepath"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/path/pathmgr"
)

// PermissionMgr permission manager for setting owner,group and mode for component's dirs and files
type PermissionMgr struct {
	CompName       string
	ConfigPathMgr  *pathmgr.ConfigPathMgr
	WorkAbsPathMgr *pathmgr.WorkAbsPathMgr
	LogPathMgr     *pathmgr.LogPathMgr
}

// SetOwnerAndGroup set owner and group for component's dirs and files
func (pm *PermissionMgr) SetOwnerAndGroup() error {
	userCfgMap, err := pm.GetUserCfgMap()
	if err != nil {
		hwlog.RunLog.Errorf("get user configuration map failed,error:%v", err)
		return errors.New("get user configuration map failed")
	}
	userCfg := userCfgMap[pm.CompName]
	for _, path := range userCfg.dirList {
		if !fileutils.IsExist(path) {
			continue
		}
		param := fileutils.SetOwnerParam{
			Path:         path,
			Uid:          userCfg.userUid,
			Gid:          userCfg.userGid,
			Recursive:    true,
			IgnoreFile:   false,
			CheckerParam: []fileutils.FileChecker{&fileutils.FileBaseChecker{}},
		}
		if err := fileutils.SetPathOwnerGroup(param); err != nil {
			hwlog.RunLog.Errorf("set path [%s] owner and group [%d] failed: %s", path, userCfg.userUid, err.Error())
			return errors.New("set path uid/gid failed")
		}
	}
	return nil
}

// SetMode set mode for component's dirs and files
func (pm *PermissionMgr) SetMode() error {
	modeCfgMap := pm.GetModeCfgMap()
	modeCfgList := modeCfgMap[pm.CompName]

	for _, modeCfg := range modeCfgList {
		switch modeCfg.types {
		case all:
			if err := pm.setAllMode(modeCfg.dirList, modeCfg.mode); err != nil {
				hwlog.RunLog.Errorf("set default mode failed: %s", err.Error())
				return errors.New("set default mode failed")
			}
		case file:
			if err := pm.setFileMode(modeCfg.dirList, modeCfg.mode); err != nil {
				return err
			}
		case dir:
			if err := pm.setDirMode(modeCfg.dirList, modeCfg.mode); err != nil {
				return err
			}
		default:
			hwlog.RunLog.Error("set mode failed, type not found")
			return errors.New("set mode failed, type not found")
		}
	}
	return nil
}

func (pm *PermissionMgr) setAllMode(dirList []string, mode os.FileMode) error {
	for _, curDir := range dirList {
		if err := fileutils.SetPathPermission(curDir, mode, true, false, &fileutils.FileBaseChecker{}); err != nil {
			if !fileutils.IsExist(curDir) {
				continue
			}
			hwlog.RunLog.Errorf("set dir [%s] mode failed: %s", curDir, err.Error())
			return err
		}
	}
	return nil
}

func (pm *PermissionMgr) setDirMode(dirList []string, mode os.FileMode) error {
	for _, curDir := range dirList {
		if !fileutils.IsExist(curDir) {
			continue
		}
		if err := fileutils.SetPathPermission(curDir, mode, true, true, &fileutils.FileBaseChecker{}); err != nil {
			hwlog.RunLog.Errorf("set dir [%s] mode failed: %s", curDir, err.Error())
			return err
		}
	}
	return nil
}

func (pm *PermissionMgr) setFileMode(fileList []string, mode os.FileMode) error {
	for _, fileType := range fileList {
		matchFiles, err := filepath.Glob(fileType)
		if err != nil {
			hwlog.RunLog.Errorf("get matched files of type [%s] failed: %s", fileType, err.Error())
			return err
		}
		for _, curFile := range matchFiles {
			if err = fileutils.SetPathPermission(curFile, mode, false, false, &fileutils.FileBaseChecker{}); err != nil {
				hwlog.RunLog.Errorf("set file [%s] mode failed: %s", curFile, err.Error())
				return err
			}
		}
	}
	return nil
}
