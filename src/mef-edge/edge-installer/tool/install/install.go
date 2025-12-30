// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package main this file for install main
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/common"
	"edge-installer/pkg/installer/install/flows"
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
	allowTmpfs       bool
	help             bool
	h                bool
)

func init() {
	flag.BoolVar(&version, "version", false, "Output the program version")
	flag.StringVar(&installRootDir, "install_dir", constants.DefaultInstallDir, "The directory for install")
	flag.StringVar(&logRootDir, "log_dir", constants.DefaultLogDir, "The directory for log")
	flag.StringVar(&logRootBackupDir, "log_backup_dir", constants.DefaultLogBackupDir,
		"The directory for backup files of log")
	flag.BoolVar(&allowTmpfs, "allow_tmpfs", false,
		"Allow the install_dir and log_backup_dir to be in the temporary file system")
	flag.BoolVar(&help, "help", false, "print the help information")
	flag.BoolVar(&h, "h", false, "print the help information")
}

func main() {
	if len(os.Args) < constants.MinArgsLen {
		fmt.Println("the required parameter is missing")
		os.Exit(constants.ProcessFailed)
	}

	flag.Parse()

	if help || h {
		flag.Usage()
		os.Exit(constants.PrintInfo)
	}
	if version {
		fmt.Printf("%s version: %s\n", BuildName, BuildVersion)
		os.Exit(constants.PrintInfo)
	}

	if utils.IsRequiredFlagNotFound() {
		fmt.Println("the required parameter is missing")
		flag.PrintDefaults()
		os.Exit(constants.ProcessFailed)
	}

	if err := checkLog(); err != nil {
		fmt.Println(err)
		os.Exit(constants.ProcessFailed)
	}

	if err := initLog(); err != nil {
		fmt.Println(err)
		os.Exit(constants.ProcessFailed)
	}

	curUser, ip, err := envutils.GetUserAndIP()
	if err != nil {
		hwlog.RunLog.Errorf("get current user or ip info failed: %s", err.Error())
		os.Exit(constants.ProcessFailed)
	}
	if err = doInstall(); err != nil {
		hwlog.RunLog.Errorf("install %s failed", constants.MEFEdgeName)
		hwlog.OpLog.Errorf("[%s@%s] install %s failed", curUser, ip, constants.MEFEdgeName)
		os.Exit(constants.ProcessFailed)
	}
	hwlog.RunLog.Infof("install %s success", constants.MEFEdgeName)
	hwlog.OpLog.Infof("[%s@%s] install %s success", curUser, ip, constants.MEFEdgeName)
}

func checkLog() error {
	if err := common.CheckLogDirs(logRootDir, logRootBackupDir, allowTmpfs); err != nil {
		return err
	}
	if err := checkExistedLogPath(logRootDir, logRootBackupDir); err != nil {
		return err
	}
	return nil
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

func doInstall() error {
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
	if !checkInstallUser() || !checkInstalled() {
		return errors.New("check before install failed")
	}

	installationPkgDir, err := path.GetInstallDir()
	if err != nil {
		hwlog.RunLog.Errorf("get install dir failed, error: %v", err)
		return errors.New("get install dir failed")
	}
	hwlog.RunLog.Info("start to install")

	pathMgr := pathmgr.NewPathMgr(installRootDir, installationPkgDir, logRootDir, logRootBackupDir)
	workAbsDir, err := pathMgr.WorkPathMgr.GetWorkAbsDir()
	if err != nil {
		return err
	}
	installFlow := flows.NewInstallFlow(pathMgr, pathmgr.NewWorkAbsPathMgr(workAbsDir), allowTmpfs)
	if err = installFlow.RunTasks(); err != nil {
		hwlog.RunLog.Errorf("install failed, %s", err.Error())
		clearWorkDir(pathMgr.WorkPathMgr.GetMefEdgeDir())
		return err
	}
	return nil
}

func checkExistedLogPath(logDir, logBackUpDir string) error {
	existedLogPaths := []string{
		pathmgr.NewLogPathMgr(logDir, logBackUpDir).GetEdgeLogDir(),
		pathmgr.NewLogPathMgr(logDir, logBackUpDir).GetEdgeLogBackupDir(),
	}
	for _, logPath := range existedLogPaths {
		logPermMgr := common.LogPermissionMgr{LogPath: logPath}
		if err := logPermMgr.CheckPermission(); err != nil {
			return fmt.Errorf("check path [%s] failed, error: %v", logPath, err)
		}
	}
	return nil
}

func checkInstallUser() bool {
	if err := envutils.CheckUserIsRoot(); err != nil {
		fmt.Println(err.Error())
		hwlog.RunLog.Errorf("check install user failed, error: %v", err)
		return false
	}
	return true
}

func checkInstalled() bool {
	softwareDir := filepath.Join(installRootDir, constants.MEFEdgeName)
	if fileutils.IsExist(softwareDir) {
		fmt.Printf("the install path [%s] already existed, please uninstall or remove it first\n", softwareDir)
		hwlog.RunLog.Errorf("the install path [%s] already existed, please uninstall or remove it first", softwareDir)
		return false
	}

	if util.IsServiceInSystemd(constants.EdgeOmServiceFile) {
		serviceFile := filepath.Join(constants.SystemdServiceDir, constants.EdgeOmServiceFile)
		entryPath, err := util.GetExecStartInService(serviceFile, constants.ModeUmask077, constants.RootUserUid)
		if err != nil {
			hwlog.RunLog.Errorf("get entry path failed, error: %v", err)
			return false
		}

		installedDirIndex := strings.LastIndex(entryPath, constants.MEFEdgeName)
		if installedDirIndex == -1 {
			hwlog.RunLog.Error("get install dir in service file failed")
			return false
		}

		installedDir := entryPath[:installedDirIndex+len(constants.MEFEdgeName)]
		if fileutils.IsExist(entryPath) {
			fmt.Printf("%s has been installed, please uninstall or remove the installed path [%s]\n",
				constants.MEFEdgeName, installedDir)
			hwlog.RunLog.Errorf("%s has been installed, please uninstall or remove the installed path [%s]",
				constants.MEFEdgeName, installedDir)
			return false
		}
	}

	serviceFiles := []string{
		constants.MefInitScriptName,
		constants.DevicePluginServiceFile,
		constants.EdgeMainServiceFile,
		constants.EdgeOmServiceFile,
		constants.EdgeCoreServiceFile,
		constants.MefEdgeTargetFile,
	}

	var remainServices []string
	for _, service := range serviceFiles {
		if util.IsServiceInSystemd(service) {
			remainServices = append(remainServices, service)
		}
	}
	if len(remainServices) == 0 {
		return true
	}

	hwlog.RunLog.Errorf("%s has been removed, but some service files remain", constants.MEFEdgeName)
	fmt.Printf("the following service files exist in [%s]\n", constants.SystemdServiceDir)
	fmt.Printf("please stop and unregister them before installation:\n\t%v\n", remainServices)
	return false
}

func clearWorkDir(workPath string) {
	if !fileutils.IsExist(workPath) {
		return
	}
	if err := fileutils.DeleteAllFileWithConfusion(workPath); err != nil {
		fmt.Printf("clean the install path failed, please remove the installed path %s\n", workPath)
		hwlog.RunLog.Errorf("clean the install path failed, please remove the installed path %s", workPath)
		return
	}
	hwlog.RunLog.Info("clean the install path success")
}
