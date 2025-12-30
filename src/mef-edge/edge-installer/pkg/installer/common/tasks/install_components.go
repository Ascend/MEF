// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tasks for prepare components' install directories
package tasks

import (
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/installer/common"
	"edge-installer/pkg/installer/common/components"
)

// InstallComponentsTask the task for install components
type InstallComponentsTask struct {
	PathMgr        *pathmgr.PathManager
	WorkAbsPathMgr *pathmgr.WorkAbsPathMgr
}

// Run components install
func (ict *InstallComponentsTask) Run() error {
	funcInfos := []common.FuncInfo{
		{Name: constants.EdgeInstaller, Function: components.NewPrepareInstaller(ict.PathMgr, ict.WorkAbsPathMgr).Run},
		{Name: constants.EdgeOm, Function: components.NewPrepareEdgeOm(ict.PathMgr, ict.WorkAbsPathMgr).Run},
		{Name: constants.EdgeMain, Function: components.NewPrepareEdgeMain(ict.PathMgr, ict.WorkAbsPathMgr).Run},
		{Name: constants.EdgeCore, Function: components.NewPrepareEdgeCore(ict.PathMgr, ict.WorkAbsPathMgr).Run},
		{Name: constants.DevicePlugin, Function: components.NewPrepareDevicePlugin(ict.PathMgr, ict.WorkAbsPathMgr).Run},
	}
	for _, info := range funcInfos {
		if err := info.Function(); err != nil {
			return fmt.Errorf("install component [%s] failed", info.Name)
		}
		hwlog.RunLog.Infof("install component [%s] success", info.Name)
	}
	return nil
}
