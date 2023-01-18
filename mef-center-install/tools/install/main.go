// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package main manages MEF cloud installation
package main

import (
	"flag"
	"fmt"
	"os"

	"huawei.com/mindx/common/hwlog"
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
	installPath            string
	help                   bool
)

func init() {
	flag.BoolVar(&version, util.VersionFlag, false, "Output the program version")
	flag.BoolVar(&installAll, util.AllInstallFlag, false, "loadImage all optional components")
	flag.BoolVar(&installSoftwareManager, util.SoftwareManagerFlag, false, "loadImage software manager")
	flag.BoolVar(&help, util.HelpFlag, false, "print the help information")
	flag.BoolVar(&help, util.HelpShortFlag, false, "print the help information")
	flag.StringVar(&logRootPath, util.LogPathFlag, "/var", "The path used to save logs")
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
	installCtlIns := install.GetSftInstallCtl(optionalComponents, installPath, logRootPath)

	if err := installCtlIns.DoInstall(); err != nil {
		return err
	}
	return nil
}

func checkPath() error {
	var err error

	if logRootPath == "" || !utils.IsExist(logRootPath) {
		return fmt.Errorf("log dir [%s] dose not exist", logRootPath)
	}

	if installPath == "" || !utils.IsExist(installPath) {
		return fmt.Errorf("install dir [%s] dose not exist", installPath)
	}

	if logRootPath, err = utils.RealDirChecker(logRootPath, true, false); err != nil {
		return fmt.Errorf("check log dir failed, error: %s", err.Error())
	}

	if installPath, err = utils.RealDirChecker(installPath, true, false); err != nil {
		return fmt.Errorf("check install dir failed, error: %s", err.Error())
	}

	return nil
}

func initLogPath(installLogPath string) error {
	if err := util.InitLogPath(installLogPath); err != nil {
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

	logPathMgr := util.InitLogDirPathMgr(logRootPath)
	installLogPath := logPathMgr.GetInstallLogPath()
	if err := common.MakeSurePath(installLogPath); err != nil {
		// install log has not initialized yet
		fmt.Printf("create log path [%s] failed\n", installLogPath)
		os.Exit(util.ErrorExitCode)
	}

	if err := initLogPath(installLogPath); err != nil {
		// install log has not initialized yet
		fmt.Println(err.Error())
		os.Exit(util.ErrorExitCode)
	}

	hwlog.OpLog.Info("start to install MEF Center")
	hwlog.RunLog.Info("--------------------Start to install MEF-Center--------------------")
	if err := doInstall(); err != nil {
		hwlog.RunLog.Errorf("install failed: %s", err.Error())
		hwlog.OpLog.Error("install MEF Center failed")
		os.Exit(1)
	}
	hwlog.RunLog.Info("--------------------Install MEF_Center success--------------------")
	hwlog.OpLog.Info("install MEF Center successfully")
}
