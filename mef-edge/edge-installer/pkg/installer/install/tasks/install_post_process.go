// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package tasks for install post process
package tasks

import (
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/installer/common/tasks"
)

// PostInstallProcessTask the task for post process after installation
type PostInstallProcessTask struct {
	tasks.PostProcessBaseTask
}

// Run post process task
func (p *PostInstallProcessTask) Run() error {
	var postFunc = []func() error{
		p.removeUpgradeBin,
		p.CreateSoftwareSymlink,
		p.UpdateMefServiceInfo,
		p.SetSoftwareDirImmutable,
	}
	for _, function := range postFunc {
		if err := function(); err != nil {
			hwlog.RunLog.Error(err)
			return err
		}
	}
	return nil
}

func (p *PostInstallProcessTask) removeUpgradeBin() error {
	workAbsDir, err := p.WorkPathMgr.GetWorkAbsDir()
	if err != nil {
		return err
	}
	upgradeBin := pathmgr.NewWorkAbsPathMgr(workAbsDir).GetUpgradeBinaryPath()
	return p.RemoveUpgradeBinByPath(upgradeBin)
}
