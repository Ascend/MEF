// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package flows this file for upgrade flow
package flows

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/installer/common"
	commonTasks "edge-installer/pkg/installer/common/tasks"
	"edge-installer/pkg/installer/upgrade/tasks"
)

type upgradeFlow struct {
	pathMgr        *pathmgr.PathManager
	workAbsPathMgr *pathmgr.WorkAbsPathMgr
}

// NewUpgradeFlow create upgrade flow instance
func NewUpgradeFlow(pathMgr *pathmgr.PathManager) common.Flow {
	return &upgradeFlow{
		pathMgr: pathMgr,
	}
}

// RunTasks run upgrade tasks
func (uf *upgradeFlow) RunTasks() error {
	checkUpgradeParam := commonTasks.CheckParamTask{
		InstallRootDir:     uf.pathMgr.GetInstallRootDir(),
		InstallationPkgDir: uf.pathMgr.GetInstallationPkgDir(),
		AllowTmpfs:         true,
	}
	if err := checkUpgradeParam.Run(); err != nil {
		return errors.New("check upgrade param task failed")
	}

	hwlog.RunLog.Info("------------------check upgrade param task success------------------")
	checkUpgradeEnvironment := tasks.CheckUpgradeEnvironmentTask{
		SoftwarePathMgr: uf.pathMgr.SoftwarePathMgr,
	}
	if err := checkUpgradeEnvironment.Run(); err != nil {
		return errors.New("check upgrade environment task failed")
	}
	fmt.Println("prepare upgrade success")

	hwlog.RunLog.Info("------------------check upgrade environment task success------------------")
	setWorkPath := tasks.SetWorkPathTask{PathMgr: uf.pathMgr}
	if err := setWorkPath.Run(); err != nil {
		return errors.New("set upgrade work path task failed")
	}
	fmt.Println("prepare software dir success")

	hwlog.RunLog.Info("------------------set upgrade work path task success------------------")

	workAbsDir, err := uf.pathMgr.WorkPathMgr.GetWorkAbsDir()
	if err != nil {
		return err
	}
	installComponents := commonTasks.InstallComponentsTask{
		PathMgr: uf.pathMgr,
		// software_temp has generated now
		WorkAbsPathMgr: pathmgr.NewWorkAbsPathMgr(workAbsDir),
	}
	if err = installComponents.Run(); err != nil {
		return errors.New("install components task failed")
	}
	hwlog.RunLog.Info("------------------install components task success------------------")
	return nil
}
