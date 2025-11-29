// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
