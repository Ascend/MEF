// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package flows for download edge installer flow
package flows

import (
	"path/filepath"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/installer/common"
	"edge-installer/pkg/installer/preupgrade/tasks"
)

type verificationInstaller struct {
	upgradeBase
	downloadPath string
}

// NewVerificationInstaller download edge installer flow
func NewVerificationInstaller(edgeDir string) common.Flow {
	const (
		progressReceived   = 25
		progressPrepareDir = 40
	)
	f := &verificationInstaller{}
	f.edgeDir = edgeDir
	f.downloadPath = constants.EdgeDownloadPath
	f.extractPath = filepath.Join(constants.UnpackPath, constants.EdgeInstaller)
	f.AddTask(tasks.LockUpgrade(), "lock upgrade", progressReceived)
	f.AddTask(tasks.PrepareDir(constants.EdgeInstaller), "prepare package dir", progressPrepareDir)
	f.AddTask(tasks.NewPrepareOnlineInstallEnv(f.downloadPath, f.extractPath, f.edgeDir),
		"check package and environment", common.ProgressSuccess)
	f.AddException(f.clearUnpackPath)
	f.AddFinal(f.clearDownloadPath, progressReceived)
	f.AddFinal(f.unlockUpgradeFlag, progressReceived)
	return f
}

func (ui *verificationInstaller) clearDownloadPath() {
	if err := fileutils.DeleteAllFileWithConfusion(ui.downloadPath); err != nil {
		hwlog.RunLog.Warnf("clean download dir [%s] failed: %v", ui.downloadPath, err)
		return
	}
	hwlog.RunLog.Info("clean download dir success")
}
