// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tasks for setting work path
package tasks

import (
	"fmt"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
)

// SetWorkPathTask the task for prepare upgrade work path
type SetWorkPathTask struct {
	PathMgr *pathmgr.PathManager
}

// Run set work path task
func (swp *SetWorkPathTask) Run() error {
	var setFunc = []func() error{
		swp.prepareInstallDir,
		swp.prepareCfgBackupDir,
	}
	for _, function := range setFunc {
		if err := function(); err != nil {
			hwlog.RunLog.Error(err)
			return err
		}
	}
	return nil
}

func (swp *SetWorkPathTask) prepareInstallDir() error {
	softwareInstallDir := swp.PathMgr.WorkPathMgr.GetUpgradeTempDir()
	if err := fileutils.DeleteAllFileWithConfusion(softwareInstallDir); err != nil {
		return fmt.Errorf("clean target software install dir failed, error: %v", err)
	}

	if err := fileutils.CreateDir(softwareInstallDir, constants.Mode755); err != nil {
		return fmt.Errorf("create target software install dir failed, error: %v", err)
	}

	hwlog.RunLog.Info("prepare install dir success")
	return nil
}
