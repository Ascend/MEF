// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tasks this file for check upgrade environment task
package tasks

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/common/tasks"
)

// CheckUpgradeEnvironmentTask the task for check upgrade environment
type CheckUpgradeEnvironmentTask struct {
	tasks.CheckEnvironmentBaseTask
	SoftwarePathMgr *pathmgr.SoftwarePathMgr
}

// Run check upgrade environment task
func (cue *CheckUpgradeEnvironmentTask) Run() error {
	var checkFunc = []func() error{
		cue.checkVersion,
		cue.checkSftPkgConsistency,
		cue.checkDiskSpace,
		cue.CheckNecessaryTools,
	}
	for _, function := range checkFunc {
		if err := function(); err != nil {
			hwlog.RunLog.Error(err)
			return err
		}
	}
	return nil
}

func (cue *CheckUpgradeEnvironmentTask) checkVersion() error {
	oldVersionPath := cue.SoftwarePathMgr.WorkPathMgr.GetVersionXmlPath()
	realOldVersionPath, err := fileutils.EvalSymlinks(oldVersionPath)
	if err != nil {
		hwlog.RunLog.Errorf("get real file path failed: %v", err)
		return errors.New("get real file path failed")
	}

	oldVersion, err := config.NewVersionXmlMgr(realOldVersionPath).GetInnerVersion()
	if err != nil {
		hwlog.RunLog.Errorf("get old inner version failed, error: %v", err)
		return errors.New("get old inner version failed")
	}

	newVersionPath := cue.SoftwarePathMgr.GetPkgVersionXmlPath()
	newVersion, err := config.NewVersionXmlMgr(newVersionPath).GetInnerVersion()
	if err != nil {
		hwlog.RunLog.Errorf("get new inner version failed, error: %v", err)
		return errors.New("get new inner version failed")
	}

	isValidVersion, err := util.IsValidVersion(oldVersion, newVersion)
	if err != nil {
		hwlog.RunLog.Errorf("compare version failed, error: %v", err)
		return err
	}
	if !isValidVersion {
		hwlog.RunLog.Errorf("upgrade version [%s] is not the previous version or the next version of "+
			"current version [%s]", newVersion, oldVersion)
		return fmt.Errorf("upgrade version [%s] is not the previous version or the next version of "+
			"current version [%s]", newVersion, oldVersion)
	}
	hwlog.RunLog.Info("check upgrade version success")
	return nil
}

func (cue *CheckUpgradeEnvironmentTask) checkSftPkgConsistency() error {
	oldVersionPath := cue.SoftwarePathMgr.WorkPathMgr.GetVersionXmlPath()
	realOldVersionPath, err := fileutils.EvalSymlinks(oldVersionPath)
	if err != nil {
		hwlog.RunLog.Errorf("get real file path failed: %v", err)
		return errors.New("get real file path failed")
	}

	oldSftPkgName, err := config.NewVersionXmlMgr(realOldVersionPath).GetSftPkgName()
	if err != nil {
		hwlog.RunLog.Errorf("get old package name failed, error: %v", err)
		return errors.New("get old package name failed")
	}

	newVersionPath := cue.SoftwarePathMgr.GetPkgVersionXmlPath()
	newSftPkgName, err := config.NewVersionXmlMgr(newVersionPath).GetSftPkgName()
	if err != nil {
		hwlog.RunLog.Errorf("get new inner package name failed, error: %v", err)
		return errors.New("get new inner package name failed")
	}

	if oldSftPkgName != newSftPkgName {
		hwlog.RunLog.Errorf("package names [old: %s new: %s] are inconsistent", oldSftPkgName, newSftPkgName)
		return errors.New("package names are inconsistent")
	}
	hwlog.RunLog.Info("check upgrade package name successfully")
	return nil
}

func (cue *CheckUpgradeEnvironmentTask) checkDiskSpace() error {
	if err := envutils.CheckDiskSpace(cue.SoftwarePathMgr.GetInstallRootDir(), constants.InstallerUpgradeMin); err != nil {
		hwlog.RunLog.Errorf("check path [%s] disk space failed, error: %v", cue.SoftwarePathMgr.GetInstallRootDir(), err)
		return errors.New("check disk space failed")
	}
	hwlog.RunLog.Info("check disk space success")
	return nil
}
