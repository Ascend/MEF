// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package flows for upgrade base struct
package flows

import (
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/common"
)

type upgradeBase struct {
	common.FlowBase
	edgeDir     string
	extractPath string
}

func (f *upgradeBase) unlockUpgradeFlag() {
	upgradeFlag := util.FlagLockInstance(constants.FlagPath, constants.ProcessFlag, constants.Upgrade)
	if err := upgradeFlag.Unlock(); err != nil {
		hwlog.RunLog.Warnf("unlock upgrade failed,%v", err)
		return
	}
	hwlog.RunLog.Info("unlock upgrade success")
}

func (f *upgradeBase) clearUnpackPath() {
	if !fileutils.IsExist(f.extractPath) {
		hwlog.RunLog.Infof("unpack package path[%s] does not exist", f.extractPath)
		return
	}
	if err := fileutils.DeleteAllFileWithConfusion(f.extractPath); err != nil {
		hwlog.RunLog.Warnf("clean unpack package path[%s] failed", f.extractPath)
		return
	}
	hwlog.RunLog.Info("clean unpack package path success")
}
