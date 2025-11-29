// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package main this file for upgrade main
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/common"
	"edge-installer/pkg/installer/upgrade/flows"
)

var (
	// BuildName the program name
	BuildName string
	// BuildVersion the program version
	BuildVersion string

	version          bool
	installRootDir   string
	logRootDir       string
	logRootBackupDir string
	config           string
	mode             string

	unpackDir string
)

func init() {
	flag.BoolVar(&version, "version", false, "Output the program version")
	flag.StringVar(&installRootDir, "install_dir", constants.DefaultInstallDir, "The directory for install")
	flag.StringVar(&logRootDir, "log_dir", constants.DefaultLogDir, "The directory for log")
	flag.StringVar(&logRootBackupDir, "log_backup_dir", constants.DefaultLogBackupDir,
		"The directory for backup files of log")
	flag.StringVar(&config, "keep_config", "all", "The reserved scope of the configuration, options: [all, min, none]")
	flag.StringVar(&mode, "mode", "other", "to distinguish if the operation is upgrading or effecting")
}

func main() {
	if len(os.Args) < constants.MinArgsLen {
		fmt.Println("the required parameter is missing")
		os.Exit(constants.ProcessFailed)
	}

	mask := syscall.Umask(constants.ModeUmask022)
	defer syscall.Umask(mask)

	flag.Parse()
	if version {
		fmt.Printf("%s version: %s\n", BuildName, BuildVersion)
		return
	}

	if err := common.CheckLogDirs(logRootDir, logRootBackupDir, true); err != nil {
		fmt.Println(err)
		os.Exit(constants.ProcessFailed)
	}

	if err := initLog(); err != nil {
		fmt.Println(err)
		os.Exit(constants.ProcessFailed)
	}

	var err error
	unpackDir, err = path.GetInstallDir()
	if err != nil {
		hwlog.RunLog.Errorf("get unpack dir failed, error: %s", err.Error())
		os.Exit(constants.ProcessFailed)
	}

	pathMgr := pathmgr.NewPathMgr(installRootDir, unpackDir, logRootDir, logRootBackupDir)

	switch mode {
	case constants.UpgradeMode:
		if err = doUpgrade(pathMgr); err != nil {
			hwlog.RunLog.Errorf("upgrade %s failed", constants.MEFEdgeName)
			os.Exit(constants.ProcessFailed)
		}
		hwlog.RunLog.Infof("upgrade %s success", constants.MEFEdgeName)
	case constants.EffectMode:
		if err = doEffect(pathMgr); err != nil {
			hwlog.RunLog.Errorf("effect %s failed", constants.MEFEdgeName)
			os.Exit(constants.ProcessFailed)
		}
	default:
		hwlog.RunLog.Error("unsupported mode")
		os.Exit(constants.ProcessFailed)
	}
}

func initLog() error {
	logPathMgr := pathmgr.NewLogPathMgr(logRootDir, logRootBackupDir)
	installLogPath := logPathMgr.GetComponentLogDir(constants.EdgeInstaller)
	installLogBackupPath := logPathMgr.GetComponentLogBackupDir(constants.EdgeInstaller)
	if err := util.InitLog(installLogPath, installLogBackupPath); err != nil {
		return fmt.Errorf("initialize log failed, error: %s", err.Error())
	}
	return nil
}

func doUpgrade(pathMgr *pathmgr.PathManager) error {
	var err error
	if installRootDir == "" {
		fmt.Println("install dir does not exist")
		hwlog.RunLog.Error("install dir does not exist")
		return errors.New("install dir does not exist")
	}
	if !filepath.IsAbs(installRootDir) {
		fmt.Println("install dir is not absolute path")
		hwlog.RunLog.Error("install dir is not absolute path")
		return errors.New("install dir is not absolute path")
	}

	hwlog.RunLog.Info("start to upgrade")

	upgradeFlow := flows.NewUpgradeFlow(pathMgr)
	if err = upgradeFlow.RunTasks(); err != nil {
		hwlog.RunLog.Errorf("upgrade failed, %s", err.Error())
		if err := clearTempDir(pathMgr); err != nil {
			hwlog.RunLog.Error("clean temp upgrade paths failed")
		}
		return err
	}

	if err = util.SetImmutable(pathMgr.WorkPathMgr.GetUpgradeTempDir()); err != nil {
		hwlog.RunLog.Warnf("set software dir immutable find errors, maybe include link file")
	}

	return nil
}

func doEffect(pathMgr *pathmgr.PathManager) error {
	hwlog.RunLog.Infof("start to effect %s", constants.MEFEdgeName)
	fmt.Printf("start to effect %s\n", constants.MEFEdgeName)

	if err := util.UnSetImmutable(pathMgr.WorkPathMgr.GetUpgradeTempDir()); err != nil {
		hwlog.RunLog.Warn("unset temp software immutable failed, maybe include link file")
	}

	upgradeFlow := flows.NewEffectFlow(pathMgr)
	if err := upgradeFlow.RunTasks(); err != nil {
		hwlog.RunLog.Errorf("upgrade failed, %s", err.Error())
		if err := clearTempDir(pathMgr); err != nil {
			hwlog.RunLog.Error("clean temp upgrade paths failed")
		}
		return err
	}
	hwlog.RunLog.Infof("effect %s success", constants.MEFEdgeName)
	fmt.Printf("effect %s success\n", constants.MEFEdgeName)
	return nil
}

func clearTempDir(pathMgr *pathmgr.PathManager) error {
	tempDirs := []string{
		pathMgr.WorkPathMgr.GetUpgradeTempDir(),
		pathMgr.ConfigPathMgr.GetConfigBackupTempDir(),
	}
	for _, dir := range tempDirs {
		if err := fileutils.DeleteAllFileWithConfusion(dir); err != nil {
			hwlog.RunLog.Errorf("clean temp path [%s] failed: %v, please remove it", dir, err)
			return err
		}
	}
	hwlog.RunLog.Info("clean temp upgrade paths success")
	return nil
}
