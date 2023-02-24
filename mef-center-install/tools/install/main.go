// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package main manages MEF cloud installation
package main

import (
	"flag"
	"fmt"
	"os"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/terminal"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/install"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

var (
	// BuildName the program name
	BuildName string
	// BuildVersion the program version
	BuildVersion string

	version                bool
	installAll             bool
	installSoftwareManager bool
	logRootPath            string
	logBackupRootPath      string
	installPath            string
	help                   bool
)

func init() {
	setFlag()
}

func setFlag() {
	flag.BoolVar(&version, util.VersionFlag, false, "Output the program version")
	flag.BoolVar(&installAll, util.AllInstallFlag, false, "loadImage all optional components")
	flag.BoolVar(&installSoftwareManager, util.SoftwareManagerFlag, false, "loadImage software manager")
	flag.BoolVar(&help, util.HelpFlag, false, "print the help information")
	flag.BoolVar(&help, util.HelpShortFlag, false, "print the help information")
	flag.StringVar(&logRootPath, util.LogPathFlag, "/var", "The path used to save logs")
	flag.StringVar(&logBackupRootPath, util.LogBackupPathFlag, "/var", "The path used to backup log files")
	flag.StringVar(&installPath, util.InstallPathFlag, "/usr/local", "The path used to install")
}

func isFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func paramOptionalComponents() []string {
	if installAll {
		return []string{
			util.SoftwareManagerFlag,
		}
	}
	var installComponents []string
	if isFlagSet(util.SoftwareManagerFlag) && installSoftwareManager {
		installComponents = append(installComponents, util.SoftwareManagerFlag)
	}

	return installComponents
}

func doInstall() error {
	optionalComponents := paramOptionalComponents()
	fullComponents := append(optionalComponents, util.GetCompulsorySlice()...)
	installCtlIns := install.GetSftInstallMgrIns(fullComponents, installPath, logRootPath, logBackupRootPath)

	if err := installCtlIns.DoInstall(); err != nil {
		return err
	}
	return nil
}

func checkPath() error {
	var err error

	if logRootPath, err = checkSinglePath(logRootPath); err != nil {
		return fmt.Errorf("check log root path failed: %s", err.Error())
	}

	if logBackupRootPath, err = checkSinglePath(logBackupRootPath); err != nil {
		return fmt.Errorf("check log back path failed: %s", err.Error())
	}

	if installPath, err = checkSinglePath(installPath); err != nil {
		return fmt.Errorf("check install path failed: %s", err.Error())
	}

	return nil
}

func checkSinglePath(path string) (string, error) {
	if path == "" || !utils.IsExist(path) {
		return "", fmt.Errorf("path [%s] does not exist", path)
	}

	absPath, err := utils.RealDirChecker(path, true, false)
	if err != nil {
		return "", err
	}

	if err = checkTmpfs(absPath); err != nil {
		return "", err
	}

	return absPath, nil
}

func checkTmpfs(path string) error {
	dev, err := common.GetFileSystem(path)
	if err != nil {
		return err
	}

	if dev == common.TmpfsDevNum {
		return fmt.Errorf("path [%s] is in tmpfs filesystem", path)
	}
	return nil
}

func initLogPath(installLogPath string, installLogBackupPath string) error {
	if err := util.InitLogPath(installLogPath, installLogBackupPath); err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(util.HelpExitCode)
	}

	if version {
		fmt.Printf("%s version: %s\n", BuildName, BuildVersion)
		os.Exit(util.VersionExitCode)
	}

	if err := checkPath(); err != nil {
		fmt.Printf("check path failed: %s\n", err.Error())
		os.Exit(util.ErrorExitCode)
	}
	fmt.Println("check path success")

	logPathMgr := util.InitLogDirPathMgr(logRootPath, logBackupRootPath)
	installLogPath := logPathMgr.GetInstallLogPath()
	installLogBackupPath := logPathMgr.GetInstallLogBackupPath()
	if err := common.MakeSurePath(installLogPath); err != nil {
		// install log has not initialized yet
		fmt.Printf("create log path [%s] failed\n", installLogPath)
		os.Exit(util.ErrorExitCode)
	}

	if err := initLogPath(installLogPath, installLogBackupPath); err != nil {
		// install log has not initialized yet
		fmt.Println(err.Error())
		os.Exit(util.ErrorExitCode)
	}
	fmt.Println("init log success")

	user, ip, err := terminal.GetLoginUserAndIP()
	if err != nil {
		hwlog.RunLog.Errorf("get current user or ip info failed: %s", err.Error())
		hwlog.OpLog.Error("install MEF Center failed: cannot get local user or ip")
		os.Exit(util.ErrorExitCode)
	}

	hwlog.OpLog.Infof("%s: %s, start to install MEF Center", ip, user)
	hwlog.RunLog.Info("--------------------Start to install MEF-Center--------------------")
	if err = doInstall(); err != nil {
		hwlog.RunLog.Errorf("install failed: %s", err.Error())
		hwlog.OpLog.Errorf("%s: %s, install MEF Center failed", ip, user)
		os.Exit(1)
	}
	hwlog.RunLog.Info("--------------------Install MEF_Center success--------------------")
	hwlog.OpLog.Infof("%s: %s, install MEF Center successfully", ip, user)
}
