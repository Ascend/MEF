// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package tasks for call upgrade task
package tasks

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/common"
)

const upgradeUmask = 022

type upgradeInstaller struct {
	installRootDir string
	mode           string
	upPath         string
	logDir         string
	logBackupDir   string
}

// UpgradeInstaller upgrade edge installer
func UpgradeInstaller(edgeDir string, mode string) common.Task {
	return &upgradeInstaller{installRootDir: filepath.Dir(edgeDir), mode: mode}
}

// Run task
func (t *upgradeInstaller) Run() error {
	return t.callUpgrade()
}

func (t *upgradeInstaller) callUpgrade() error {
	if err := t.initUpgradePara(); err != nil {
		return fmt.Errorf("init upgrade para failed, error: %v", err)
	}
	switch t.mode {
	case constants.DefaultMode:
		t.mode = constants.UpgradeMode
		if err := t.runUpgrade(); err != nil {
			return fmt.Errorf("call bin for upgrade failed, error: %v", err)
		}
		t.mode = constants.EffectMode
		if err := t.runUpgrade(); err != nil {
			return fmt.Errorf("call bin for effect failed, error: %v", err)
		}
	case constants.Upgrade, constants.EffectMode:
		if err := t.runUpgrade(); err != nil {
			return fmt.Errorf("call upgrade bin failed, error: %v", err)
		}
	default:
		return errors.New("unknown call mode for upgrade installer")
	}
	return nil
}

func (t *upgradeInstaller) initUpgradePara() error {
	logRootDir, err := path.GetLogRootDir(t.installRootDir)
	if err != nil {
		return fmt.Errorf("get log root dir failed, error: %v", err)
	}
	logBackupRootDir, err := path.GetLogBackupRootDir(t.installRootDir)
	if err != nil {
		return fmt.Errorf("get log backup root dir failed, error: %v", err)
	}
	t.logDir = logRootDir
	t.logBackupDir = logBackupRootDir
	return nil
}

func (t *upgradeInstaller) runUpgrade() error {
	if err := util.CheckNecessaryCommands(); err != nil {
		return errors.New("check necessary commands failed")
	}

	t.upPath = constants.UpgradePath
	if t.mode == constants.EffectMode {
		t.upPath = pathmgr.NewWorkPathMgr(t.installRootDir).GetUpgradeTempBinaryPath()
	}

	if err := t.checkOwnerAndPermission(t.upPath, upgradeUmask, constants.UpgradeUid); err != nil {
		return fmt.Errorf("upgrade file check invalid: %v", err)
	}

	return envutils.RunCommandWithOsStdout(t.upPath, envutils.DefCmdTimeoutSec,
		fmt.Sprintf("--install_dir=%s", t.installRootDir),
		fmt.Sprintf("--log_dir=%s", t.logDir),
		fmt.Sprintf("--log_backup_dir=%s", t.logBackupDir),
		"--keep_config=all",
		fmt.Sprintf("--mode=%s", t.mode))
}

func (t *upgradeInstaller) checkOwnerAndPermission(verifyPath string, modeUmask os.FileMode, userId uint32) error {
	ownerChecker := fileutils.NewFileOwnerChecker(true, false, userId, userId)
	modeChecker := fileutils.NewFileModeChecker(true, modeUmask, false, false)
	linkChecker := fileutils.NewFileLinkChecker(false)
	ownerChecker.SetNext(modeChecker)
	ownerChecker.SetNext(linkChecker)

	file, err := os.OpenFile(verifyPath, os.O_RDONLY, constants.Mode400)
	if err != nil {
		return fmt.Errorf("open file %s failed", verifyPath)
	}
	defer func() {
		if err = file.Close(); err != nil {
			hwlog.RunLog.Errorf("close file %s failed", verifyPath)
		}
	}()

	if err = ownerChecker.Check(file, verifyPath); err != nil {
		hwlog.RunLog.Errorf("check file %s failed, error: %v", verifyPath, err)
		return errors.New("check file failed")
	}
	return nil
}
