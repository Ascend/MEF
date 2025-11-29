// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
	CmsPath     string
	CrlPath     string
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
	f.AddTask(tasks.NewCheckOfflineEdgeInstallerEnv(param.TarPath, param.CmsPath, param.CrlPath, f.extractPath,
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
