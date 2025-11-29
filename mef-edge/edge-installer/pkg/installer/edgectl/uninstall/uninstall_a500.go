// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_A500

package uninstall

import (
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
)

func (pu processUninstallTask) Run() error {
	var setFunc = []func() error{
		pu.unsetImmutable,
		pu.removeService,
		pu.removeExternalFiles,
		pu.removeContainer,
		pu.restorePrimalDockerConfig,
		pu.removeInstallDir,
	}
	for _, function := range setFunc {
		if err := function(); err != nil {
			return err
		}
	}
	return nil
}

// The container configuration restoration failure does not affect the uninstallation result.
func (pu processUninstallTask) restorePrimalDockerConfig() error {
	primalDockerServiceBackupPath := pu.configPathMgr.GetDockerBackupPath()
	dockerRestoreShPath := pu.workPathMgr.GetDockerRestoreShPath()
	installerConfigDir := pu.configPathMgr.GetCompConfigDir(constants.EdgeInstaller)

	if !fileutils.IsExist(primalDockerServiceBackupPath) || !fileutils.IsExist(dockerRestoreShPath) {
		hwlog.RunLog.Info("no need to restore docker")
		return nil
	}

	realDockerRestoreSh, err := fileutils.EvalSymlinks(dockerRestoreShPath)
	if err != nil {
		hwlog.RunLog.Warnf("get docker restore sh failed: %v", err)
		return nil
	}
	out, err := envutils.RunCommand(realDockerRestoreSh, envutils.DefCmdTimeoutSec, installerConfigDir)
	if err != nil {
		hwlog.RunLog.Warnf("execute restore docker cmd failed: output: %s, err:%v", out, err)
		return nil
	}

	hwlog.RunLog.Infof("restore docker success: output: %s", out)
	return nil
}
