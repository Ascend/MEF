// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package uninstall this file for uninstall
package uninstall

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/common"
)

// FlowUninstall uninstall flow
type FlowUninstall struct {
	workPathMgr   *pathmgr.WorkPathMgr
	configPathMgr *pathmgr.ConfigPathMgr
}

type processUninstallTask struct {
	workPathMgr   *pathmgr.WorkPathMgr
	configPathMgr *pathmgr.ConfigPathMgr
}

// NewFlowUninstall create uninstall flow instance
func NewFlowUninstall(workPathMgr *pathmgr.WorkPathMgr, configPathMgr *pathmgr.ConfigPathMgr) *FlowUninstall {
	return &FlowUninstall{
		workPathMgr:   workPathMgr,
		configPathMgr: configPathMgr,
	}
}

// RunTasks run uninstall tasks
func (fu FlowUninstall) RunTasks() error {
	fmt.Println("Uninstalling might take a few minutes, please wait...")
	processUninstall := processUninstallTask{
		workPathMgr:   fu.workPathMgr,
		configPathMgr: fu.configPathMgr,
	}
	if err := processUninstall.Run(); err != nil {
		hwlog.RunLog.Errorf("process uninstall task failed, error: %v", err)
		return err
	}

	hwlog.RunLog.Info("------------------process uninstall task success------------------")
	return nil
}

func (pu processUninstallTask) unsetImmutable() error {
	if err := util.UnSetImmutable(pu.workPathMgr.GetMefEdgeDir()); err != nil {
		hwlog.RunLog.Warnf("unset edge dir [%s] immutable find errors, maybe include link file",
			pu.workPathMgr.GetMefEdgeDir())
	}
	return nil
}

func (pu processUninstallTask) removeService() error {
	mgr := common.NewComponentMgr(pu.workPathMgr.GetInstallRootDir())
	if err := mgr.StopAll(); err != nil {
		fmt.Println("stop all services failed, please stop manually")
		hwlog.RunLog.Error(err)
		return errors.New("stop all services failed")
	}
	if err := mgr.UnregisterAllServices(); err != nil {
		fmt.Println("remove all services failed, please remove manually")
		hwlog.RunLog.Error(err)
		return errors.New("remove all services failed")
	}
	hwlog.RunLog.Info("stop and remove all services success")
	return nil
}

func (pu processUninstallTask) removeExternalFiles() error {
	externalPaths := []string{
		constants.PreUpgradePath,
	}
	for _, externalPath := range externalPaths {
		if !fileutils.IsExist(externalPath) {
			continue
		}
		if err := fileutils.DeleteAllFileWithConfusion(externalPath); err != nil {
			hwlog.RunLog.Errorf("remove [%s] failed, error: %v", externalPath, err)
			return fmt.Errorf("remove [%s] failed", externalPath)
		}
	}

	hwlog.RunLog.Info("remove external files success")
	return nil
}

func (pu processUninstallTask) removeInstallDir() error {
	edgeDir := pu.workPathMgr.GetMefEdgeDir()
	hwlog.RunLog.Infof("All keys in %s are about to be destroyed.", edgeDir)
	if err := fileutils.DeleteAllFileWithConfusion(edgeDir); err != nil {
		fmt.Println("remove the installed dir failed, please remove the installed dir manually")
		hwlog.RunLog.Errorf("remove software dir [%s] failed, error: %v", edgeDir, err)
		return fmt.Errorf("remove software dir [%s] failed", edgeDir)
	}
	hwlog.RunLog.Info("remove software dir success")
	return nil
}

func (pu processUninstallTask) removeContainer() error {
	return util.RemoveContainer()
}
