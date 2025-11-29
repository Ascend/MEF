// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package flows for online upgrade edge installer flow
package flows

import (
	"path/filepath"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/installer/common"
	"edge-installer/pkg/installer/preupgrade/tasks"
)

type onlineUpgradeInstaller struct {
	upgradeBase
	pkgPath string
}

// OnlineUpgradeInstaller online effect edge installer
func OnlineUpgradeInstaller(edgeDir string) common.Flow {
	const progressPrepare = 10
	f := &onlineUpgradeInstaller{}
	f.edgeDir = edgeDir
	f.extractPath = filepath.Join(constants.UnpackPath, constants.EdgeInstaller)
	f.AddTask(tasks.LockUpgrade(), "lock upgrade", progressPrepare)
	f.AddTask(tasks.UpgradeInstaller(f.edgeDir, constants.UpgradeMode), "upgrade", common.ProgressSuccess)
	f.AddException(f.onlineUpgradeOpLogFailed)
	f.AddFinal(f.clearUnpackPath, progressPrepare)
	f.AddFinal(f.unlockUpgradeFlag, progressPrepare)
	f.AddFinal(f.onlineUpgradeOpLogOk, common.ProgressSuccess)
	// the software will be restarted during the effect process, so the effect process is placed at the last step.
	f.AddFinal(f.effect, common.ProgressSuccess)
	return f
}

func (f *onlineUpgradeInstaller) onlineUpgradeOpLogOk() {
	hwlog.OpLog.Infof("[%s@%s][%s upgrade %s][success]",
		config.NetMgr.NetType, config.NetMgr.IP, config.NetMgr.NetType, constants.MEFEdgeName)
}

func (f *onlineUpgradeInstaller) onlineUpgradeOpLogFailed() {
	hwlog.OpLog.Errorf("[%s@%s][%s upgrade %s][failed]",
		config.NetMgr.NetType, config.NetMgr.IP, config.NetMgr.NetType, constants.MEFEdgeName)
}

func (f *onlineUpgradeInstaller) effect() {
	effect := tasks.EffectInstaller(f.edgeDir)
	if err := effect.Run(); err != nil {
		hwlog.RunLog.Warnf("online upgrade failed, error: %v", err)
	}
}
