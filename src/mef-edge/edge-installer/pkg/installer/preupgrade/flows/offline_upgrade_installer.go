// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package flows for offline upgrade edge installer
package flows

import (
	"path/filepath"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/installer/common"
	"edge-installer/pkg/installer/preupgrade/posts"
	"edge-installer/pkg/installer/preupgrade/tasks"
)

type offlineUpgradeInstaller struct {
	upgradeBase
}

// OfflineUpgradeInstallerParam offline upgrade edge installer parameter
type OfflineUpgradeInstallerParam struct {
	TarPath     string
	EdgeDir     string
	DelayEffect bool
}

// OfflineUpgradeInstaller offline upgrade edge-installer flow
func OfflineUpgradeInstaller(param OfflineUpgradeInstallerParam) common.Flow {
	const (
		progressPrepare = 10
		progressCheck   = 40
		processUpgrade  = 70
	)
	f := &offlineUpgradeInstaller{}
	f.edgeDir = param.EdgeDir
	f.extractPath = filepath.Join(constants.UnpackPath, constants.EdgeInstaller)
	f.AddTask(tasks.NewCheckOfflineEdgeInstallerEnv(param.TarPath, f.extractPath,
		param.EdgeDir), "check package and environment", progressCheck)
	if param.DelayEffect {
		f.AddTask(tasks.UpgradeInstaller(f.edgeDir, constants.Upgrade), "upgrade", common.ProgressSuccess)
	} else {
		f.AddTask(tasks.UpgradeInstaller(f.edgeDir, constants.Upgrade), "upgrade", processUpgrade)
		f.AddTask(tasks.EffectInstaller(f.edgeDir), "effect", common.ProgressSuccess)
	}
	f.AddPost(posts.PrintProgress)
	f.AddFinal(f.clearUnpackPath, progressPrepare)
	return f
}
