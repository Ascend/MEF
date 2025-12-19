// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tasks this file for check install environment task
package tasks

import (
	"errors"
	"fmt"
	"strings"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/common/tasks"
)

// CheckInstallEnvironmentTask the task for check installation environment
type CheckInstallEnvironmentTask struct {
	tasks.CheckEnvironmentBaseTask
	InstallRootDir string
	LogPathMgr     *pathmgr.LogPathMgr
}

// Run check environment task
func (cie *CheckInstallEnvironmentTask) Run() error {
	var checkFunc = []func() error{
		cie.checkDiskSpace,
		cie.CheckNecessaryTools,
		cie.checkNecessaryCommands,
	}
	for _, function := range checkFunc {
		if err := function(); err != nil {
			hwlog.RunLog.Error(err)
			return err
		}
	}
	return nil
}

func (cie *CheckInstallEnvironmentTask) checkDiskSpace() error {
	var err error
	devMap := make(map[uint64]uint64)
	checkMap := []struct {
		path          string
		needDiskSpace uint64
	}{
		{cie.InstallRootDir, constants.InstallMinDiskSpace},
		{cie.LogPathMgr.GetEdgeLogDir(), constants.LogMinDiskSpace},
		{cie.LogPathMgr.GetEdgeLogBackupDir(), constants.LogBackupMinDiskSpace},
	}
	for _, check := range checkMap {
		devMap, err = cie.checkDevDiskSpace(check.path, check.needDiskSpace, devMap)
		if err != nil {
			if check.path != cie.InstallRootDir && strings.Contains(err.Error(), "no enough space") {
				hwlog.RunLog.Warnf("check path [%s] disk space failed: %v", check.path, err)
				continue
			}
			hwlog.RunLog.Errorf("check path [%s] disk space failed: %v", check.path, err)
			return errors.New("check path disk space failed")
		}
	}
	hwlog.RunLog.Info("check disk space success")
	return nil
}

func (cie *CheckInstallEnvironmentTask) checkDevDiskSpace(path string, space uint64,
	devMap map[uint64]uint64) (map[uint64]uint64, error) {
	if devMap == nil {
		return nil, errors.New("devMap is nil")
	}

	devInfo, err := envutils.GetFileDevNum(path)
	if err != nil {
		return nil, fmt.Errorf("get path [%s] dev num failed: %v", path, err)
	}

	if _, existed := devMap[devInfo]; existed {
		devMap[devInfo] += space
	} else {
		devMap[devInfo] = space
	}

	if err = envutils.CheckDiskSpace(path, devMap[devInfo]); err != nil {
		return devMap, fmt.Errorf("check path [%s] disk space failed: %v", path, err)
	}
	return devMap, nil
}

func (cie *CheckInstallEnvironmentTask) checkNecessaryCommands() error {
	if err := util.CheckNecessaryCommands(); err != nil {
		fmt.Println(err)
		return errors.New("check necessary commands failed")
	}
	hwlog.RunLog.Info("check necessary commands success")
	return nil
}
