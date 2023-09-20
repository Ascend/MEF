// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package main manages MEF cloud installation
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"huawei.com/mindx/common/envutils"
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

	version           bool
	logRootPath       string
	logBackupRootPath string
	installPath       string
	help              bool
)

func init() {
	setFlag()
}

func setFlag() {
	flag.BoolVar(&version, util.VersionFlag, false, "Output the program version")
	flag.BoolVar(&help, util.HelpFlag, false, "print the help information")
	flag.BoolVar(&help, util.HelpShortFlag, false, "print the help information")
	flag.StringVar(&logRootPath, util.LogPathFlag, "/var", "The path used to save logs")
	flag.StringVar(&logBackupRootPath, util.LogBackupPathFlag, "/var", "The path used to backup log files")
	flag.StringVar(&installPath, util.InstallPathFlag, "/usr/local", "The path used to install")
}

func doInstall() error {
	fullComponents := util.GetCompulsorySlice()
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

	if err = checkInsideLogPath(logRootPath, logBackupRootPath); err != nil {
		return err
	}

	if installPath, err = checkSinglePath(installPath); err != nil {
		return fmt.Errorf("check install path failed: %s", err.Error())
	}

	return nil
}

func checkSinglePath(singlePath string) (string, error) {
	if singlePath == "" || !utils.IsExist(singlePath) {
		return "", fmt.Errorf("path [%s] does not exist", singlePath)
	}

	if !path.IsAbs(singlePath) {
		return "", fmt.Errorf("path [%s] is not abs path", singlePath)
	}

	absPath, err := utils.RealDirChecker(singlePath, true, false)
	if err != nil {
		return "", err
	}

	if err = checkTmpfs(absPath); err != nil {
		return "", err
	}

	return absPath, nil
}

func checkInsideLogPath(logPath, logBackUpPath string) error {
	logPathMgr := util.InitLogDirPathMgr(logPath, logBackUpPath)
	logModuleDir := logPathMgr.GetModuleLogPath()

	if utils.IsExist(logModuleDir) {
		if _, err := utils.RealDirChecker(logModuleDir, false, false); err != nil {
			return fmt.Errorf("check log dir failed: %s", err.Error())
		}
		if err := filepath.Walk(logModuleDir, checkLogPath); err != nil {
			return fmt.Errorf("check log dir failed: %s", err.Error())
		}
	}

	logBackModuleDir := logPathMgr.GetModuleLogBackupPath()
	if utils.IsExist(logBackModuleDir) {
		if _, err := utils.RealDirChecker(logBackModuleDir, false, false); err != nil {
			return fmt.Errorf("check log dir failed: %s", err.Error())
		}
		if err := filepath.Walk(logBackModuleDir, checkLogPath); err != nil {
			return fmt.Errorf("check log back up dir failed: %s", err.Error())
		}
	}

	return nil
}

func checkLogPath(path string, info fs.FileInfo, err error) error {
	if err != nil {
		return err
	}

	const (
		groupWriteIndex = 5
		otherWriteIndex = 8
		permLength      = 10
	)
	perm := info.Mode().Perm().String()
	if len(perm) != permLength {
		return fmt.Errorf("permission not right %v %v", path, perm)
	}
	for index, char := range perm {
		if (index == groupWriteIndex || index == otherWriteIndex) && char == 'w' {
			return fmt.Errorf("write permission not right %v %v", path, perm)
		}
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("can not get stat %v", path)
	}
	if int(stat.Uid) == util.RootUid && int(stat.Gid) == util.RootGid {
		return nil
	}

	mefUid, mefGid, err := util.GetMefId()
	if err != nil {
		if strings.Contains(err.Error(), "unknown user") || strings.Contains(err.Error(), "unknown group") {
			return fmt.Errorf("owner not right %v uid=%v gid=%v", path, int(stat.Uid), int(stat.Gid))
		}

		return fmt.Errorf("get mef uid failed: %s", err.Error())
	}
	if stat.Uid == mefUid && stat.Gid == mefGid {
		return nil
	}
	return fmt.Errorf("owner not right %v uid=%v gid=%v", path, int(stat.Uid), int(stat.Gid))
}

func checkTmpfs(path string) error {
	isInTmpfs, err := envutils.IsInTmpfs(path)
	if err != nil {
		return err
	}
	if isInTmpfs {
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
	errCode := preCheck()
	if errCode != 0 {
		os.Exit(errCode)
	}
	user, ip, err := envutils.GetUserAndIP()
	if err != nil {
		hwlog.RunLog.Errorf("get current user or ip info failed: %s", err.Error())
		os.Exit(util.ErrorExitCode)
	}

	err = envutils.GetFlock(util.MefCenterLock).Lock("install")
	if err != nil {
		fmt.Println("the last installation is not complete")
		hwlog.RunLog.Error("install MEF Center failed: " +
			"the last installation is not complete, has not been unlocked yet")
		hwlog.OpLog.Errorf("[%s@%s] install MEF Center failed", user, ip)
		os.Exit(util.ErrorExitCode)
	}
	defer envutils.GetFlock(util.MefCenterLock).Unlock()

	hwlog.OpLog.Infof("[%s@%s] start to install MEF Center", user, ip)
	hwlog.RunLog.Info("--------------------Start to install MEF-Center--------------------")
	if err = doInstall(); err != nil {
		hwlog.RunLog.Errorf("install failed: %s", err.Error())
		hwlog.OpLog.Errorf("[%s@%s] install MEF Center failed", user, ip)
		os.Exit(util.ErrorExitCode)
	}
	hwlog.RunLog.Info("--------------------Install MEF_Center success--------------------")
	hwlog.OpLog.Infof("[%s@%s] install MEF Center successfully", user, ip)
}

func preCheck() int {
	flag.Parse()

	if help {
		flag.Usage()
		return util.HelpExitCode
	}

	if version {
		fmt.Printf("%s version: %s\n", BuildName, BuildVersion)
		return util.VersionExitCode
	}
	if utils.IsRequiredFlagNotFound() {
		fmt.Println("the required parameter is missing")
		flag.PrintDefaults()
		return util.ErrorExitCode
	}

	if err := checkPath(); err != nil {
		fmt.Printf("check path failed: %s\n", err.Error())
		return util.ErrorExitCode
	}
	fmt.Println("check path success")

	logPathMgr := util.InitLogDirPathMgr(logRootPath, logBackupRootPath)
	installLogPath := logPathMgr.GetInstallLogPath()
	installLogBackupPath := logPathMgr.GetInstallLogBackupPath()
	if err := common.MakeSurePath(installLogPath); err != nil {
		// install log has not initialized yet
		fmt.Printf("create log path [%s] failed\n", installLogPath)
		return util.ErrorExitCode
	}

	if err := initLogPath(installLogPath, installLogBackupPath); err != nil {
		// install log has not initialized yet
		fmt.Println(err.Error())
		return util.ErrorExitCode
	}
	fmt.Println("init log success")
	return 0
}
