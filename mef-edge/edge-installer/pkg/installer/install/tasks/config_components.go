// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tasks for prepare components' config directories
package tasks

import (
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/installer/common"
	"edge-installer/pkg/installer/common/components"
)

// ConfigComponentsTask the task for components config
type ConfigComponentsTask struct {
	PathMgr        *pathmgr.PathManager
	WorkAbsPathMgr *pathmgr.WorkAbsPathMgr
}

// Run components config
func (cct *ConfigComponentsTask) Run() error {
	funcInfos := []common.FuncInfo{
		{Name: constants.EdgeInstaller,
			Function: components.NewPrepareInstaller(cct.PathMgr, cct.WorkAbsPathMgr).PrepareCfgDir},
		{Name: constants.EdgeOm, Function: components.NewPrepareEdgeOm(cct.PathMgr, cct.WorkAbsPathMgr).PrepareCfgDir},
		{Name: constants.EdgeMain, Function: components.NewPrepareEdgeMain(cct.PathMgr, cct.WorkAbsPathMgr).PrepareCfgDir},
		{Name: constants.EdgeCore, Function: components.NewPrepareEdgeCore(cct.PathMgr, cct.WorkAbsPathMgr).PrepareCfgDir},
	}
	for _, info := range funcInfos {
		if err := info.Function(); err != nil {
			hwlog.RunLog.Errorf("prepare [%s] config directories failed: %v", info.Name, err)
			return fmt.Errorf("prepare [%s] config directories failed", info.Name)
		}
		hwlog.RunLog.Infof("prepare [%s] config directories success", info.Name)
	}
	return nil
}
