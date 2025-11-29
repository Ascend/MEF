// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package tasks this file for check parameters task
package tasks

import (
	"fmt"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/installer/common"
)

// CheckParamTask the task for check parameters
type CheckParamTask struct {
	InstallRootDir     string
	InstallationPkgDir string
	AllowTmpfs         bool
}

// Run check parameters task
func (cpt *CheckParamTask) Run() error {
	var checkFunc = []func() error{
		cpt.checkInstallationPkgDir,
		cpt.checkInstallRootDir,
	}
	for _, function := range checkFunc {
		if err := function(); err != nil {
			hwlog.RunLog.Error(err)
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (cpt *CheckParamTask) checkInstallationPkgDir() error {
	if _, err := fileutils.RealDirCheck(cpt.InstallationPkgDir, true, false); err != nil {
		return fmt.Errorf("check install package dir [%s] failed, error: %v", cpt.InstallationPkgDir, err)
	}
	hwlog.RunLog.Info("check install package dir success")
	return nil
}

func (cpt *CheckParamTask) checkInstallRootDir() error {
	if err := common.CheckDir(cpt.InstallRootDir, constants.InstallDirName); err != nil {
		return err
	}
	if err := common.CheckInTmpfs(cpt.InstallRootDir, cpt.AllowTmpfs); err != nil {
		return err
	}
	hwlog.RunLog.Info("check install root dir success")
	return nil
}
