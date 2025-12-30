// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
