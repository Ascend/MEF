// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
