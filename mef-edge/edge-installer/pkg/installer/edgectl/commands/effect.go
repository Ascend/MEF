// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package commands

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/edgectl/common"
)

type effectCmd struct {
}

// EffectCmd  edge control command effect mef-edge
func EffectCmd() common.Command {
	return &effectCmd{}
}

func (cmd *effectCmd) Name() string {
	return common.Effect
}

func (cmd *effectCmd) Description() string {
	return common.EffectDesc
}

// BindFlag command flag binding
func (cmd *effectCmd) BindFlag() bool {
	return false
}

// LockFlag command lock flag
func (cmd *effectCmd) LockFlag() bool {
	return true
}

// Execute execute command
func (cmd *effectCmd) Execute(ctx *common.Context) error {
	if ctx == nil {
		hwlog.RunLog.Error("ctx is nil")
		return errors.New("ctx is nil")
	}
	if !fileutils.IsExist(ctx.WorkPathMgr.GetUpgradeTempDir()) {
		fmt.Println("no software to effect")
		hwlog.RunLog.Errorf("no software is to effect")
		return errors.New("no software is to effect")
	}
	upgradeVersionPath := ctx.WorkPathMgr.GetUpgradeTempVersionXmlPath()
	upgradeMgr := config.NewVersionXmlMgr(upgradeVersionPath)
	upgradeVersion, err := upgradeMgr.GetVersion()
	if err != nil {
		hwlog.RunLog.Errorf("get old inner version failed, error: %v", err)
		return err
	}
	upgradeBinPath := ctx.WorkPathMgr.GetUpgradeTempBinaryPath()
	upgradeBinDir, err := fileutils.CheckOwnerAndPermission(upgradeBinPath, constants.UpgradeUmask,
		constants.UpgradeUid)
	if err != nil {
		return errors.New("check effect bin file failed")
	}
	installRootDir := ctx.WorkPathMgr.GetInstallRootDir()
	logRootDir, err := path.GetLogRootDir(installRootDir)
	if err != nil {
		return fmt.Errorf("get log root dir failed, error: %v", err)
	}
	logBackupRootDir, err := path.GetLogBackupRootDir(installRootDir)
	if err != nil {
		return fmt.Errorf("get log backup root dir failed, error: %v", err)
	}

	softwareDir, err := pathmgr.GetTargetInstallDir(ctx.WorkPathMgr.GetInstallRootDir())
	if err != nil {
		return fmt.Errorf("get software source dir failed, error: %s", err.Error())
	}
	if err = util.UnSetImmutable(softwareDir); err != nil {
		hwlog.RunLog.Warn("unset temp software immutable failed, maybe include link file")
	}
	if err = config.EffectToOldestVersionSmooth(upgradeVersion, installRootDir); err != nil {
		hwlog.RunLog.Errorf("smooth config file for old version failed, error: %v", err)
		return errors.New("smooth config file for old version failed")
	}
	if err = envutils.RunCommandWithOsStdout(upgradeBinDir, envutils.DefCmdTimeoutSec,
		fmt.Sprintf("--install_dir=%s", installRootDir), fmt.Sprintf("--log_dir=%s", logRootDir),
		fmt.Sprintf("--log_backup_dir=%s", logBackupRootDir), "--keep_config=all", "--mode=effect"); err != nil {
		return fmt.Errorf("call upgrade bin failed, error: %v", err)
	}
	return nil
}

// PrintOpLogOk print operation success log
func (cmd *effectCmd) PrintOpLogOk(user, ip string) {
	hwlog.OpLog.Infof("[%s@%s] effect software success", user, ip)
}

// PrintOpLogFail print operation fail log
func (cmd *effectCmd) PrintOpLogFail(user, ip string) {
	hwlog.OpLog.Errorf("[%s@%s] effect software failed", user, ip)
}
