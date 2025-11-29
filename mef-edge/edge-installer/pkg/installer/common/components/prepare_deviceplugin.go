// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package components this file for prepare device plugin
package components

import (
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
)

// PrepareDevicePlugin for prepare device plugin
type PrepareDevicePlugin struct {
	PrepareCompBase
}

// NewPrepareDevicePlugin create prepare device plugin instance
func NewPrepareDevicePlugin(pathMgr *pathmgr.PathManager, workAbsPathMgr *pathmgr.WorkAbsPathMgr) *PrepareDevicePlugin {
	return &PrepareDevicePlugin{
		PrepareCompBase: PrepareCompBase{
			CompName:       constants.DevicePlugin,
			PathManager:    pathMgr,
			WorkAbsPathMgr: workAbsPathMgr,
		},
	}
}

// Run prepare device plugin
func (pdp *PrepareDevicePlugin) Run() error {
	var preFunc = []func() error{
		pdp.prepareSoftwareDir,
		pdp.prepareLogDirs,
		pdp.prepareLogLinks,
		pdp.setOwnerAndMode,
	}
	for _, function := range preFunc {
		if err := function(); err != nil {
			hwlog.RunLog.Error(err)
			return err
		}
	}
	return nil
}
