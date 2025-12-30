// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package flows this file for install flow
package flows

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/installer/common"
	commonTasks "edge-installer/pkg/installer/common/tasks"
	"edge-installer/pkg/installer/install/tasks"
)

// FlowInstall install flow
type FlowInstall struct {
	pathMgr        *pathmgr.PathManager
	workAbsPathMgr *pathmgr.WorkAbsPathMgr
	allowTmpfs     bool
}

// NewInstallFlow create install flow instance
func NewInstallFlow(pathMgr *pathmgr.PathManager, workAbsPathMgr *pathmgr.WorkAbsPathMgr,
	allowTmpfs bool) *FlowInstall {
	return &FlowInstall{
		pathMgr:        pathMgr,
		workAbsPathMgr: workAbsPathMgr,
		allowTmpfs:     allowTmpfs,
	}
}

func (fi *FlowInstall) checkTask() error {
	checkInstallParam := commonTasks.CheckParamTask{
		InstallRootDir:     fi.pathMgr.GetInstallRootDir(),
		InstallationPkgDir: fi.pathMgr.GetInstallationPkgDir(),
		AllowTmpfs:         fi.allowTmpfs,
	}
	if err := checkInstallParam.Run(); err != nil {
		return errors.New("check install param task failed")
	}
	fmt.Println("check install parameters success")
	hwlog.RunLog.Info("------------------check install param task success------------------")

	checkInstallEnvironment := tasks.CheckInstallEnvironmentTask{
		InstallRootDir: fi.pathMgr.GetInstallRootDir(), LogPathMgr: fi.pathMgr.LogPathMgr}
	if err := checkInstallEnvironment.Run(); err != nil {
		return errors.New("check install environment task failed")
	}

	fmt.Println("check install environment success")
	hwlog.RunLog.Info("------------------check install environment task success------------------")
	return nil
}

// RunTasks run install tasks
func (fi *FlowInstall) RunTasks() error {
	if err := fi.checkTask(); err != nil {
		return err
	}

	addUserAccountTask := tasks.AddUserAccountTask{}
	if err := addUserAccountTask.Run(); err != nil {
		return errors.New("add user account task failed")
	}

	hwlog.RunLog.Info("------------------add user account task success------------------")
	setWorkPathTask := tasks.SetWorkPathTask{PathMgr: fi.pathMgr}
	if err := setWorkPathTask.Run(); err != nil {
		return errors.New("set work path task failed")
	}

	fmt.Println("prepare install and log root directories success")
	fmt.Println("installing...")
	hwlog.RunLog.Info("------------------set work path task success------------------")

	generateCertsTask, err := common.NewGenerateCertsTask(fi.pathMgr.GetInstallRootDir())
	if err != nil {
		return errors.New("get generate certs task failed")
	}

	configComponentsTask := tasks.ConfigComponentsTask{PathMgr: fi.pathMgr, WorkAbsPathMgr: fi.workAbsPathMgr}
	installComponentsTask := commonTasks.InstallComponentsTask{PathMgr: fi.pathMgr, WorkAbsPathMgr: fi.workAbsPathMgr}
	setSystemInfoTask := commonTasks.SetSystemInfoTask{
		ConfigDir:     fi.pathMgr.ConfigPathMgr.GetConfigDir(),
		ConfigPathMgr: fi.pathMgr.SoftwarePathMgr.ConfigPathMgr,
		LogPathMgr:    fi.pathMgr.LogPathMgr,
	}
	postProcessTask := tasks.PostInstallProcessTask{
		PostProcessBaseTask: commonTasks.PostProcessBaseTask{
			WorkPathMgr: fi.pathMgr.SoftwarePathMgr.WorkPathMgr,
			LogPathMgr:  fi.pathMgr.LogPathMgr,
		},
	}

	taskInfos := []common.FuncInfo{
		{Name: "config components task", Function: configComponentsTask.Run},
		{Name: "install components task", Function: installComponentsTask.Run},
		{Name: "generate certs task", Function: generateCertsTask.Run},
		{Name: "set system info task", Function: setSystemInfoTask.Run},
		{Name: "install post process task", Function: postProcessTask.Run},
	}
	for _, task := range taskInfos {
		if err := task.Function(); err != nil {
			return fmt.Errorf("%s failed", task.Name)
		}
		hwlog.RunLog.Infof("------------------%s success------------------", task.Name)
	}
	return nil
}
