// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package main manages MEF cloud start, stop and restart
package main

import (
	"fmt"
	"os"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/mef-center-install/pkg/control"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

type upgradeController struct {
	installParam *util.InstallParamJsonTemplate
	logPath      string
	logBackPath  string
}

func main() {
	installParam, err := util.GetInstallInfo()
	if err != nil {
		fmt.Printf("get info from install-param.json failed:%s\n", err.Error())
		os.Exit(util.ErrorExitCode)
	}

	if err = initLog(installParam); err != nil {
		fmt.Println(err.Error())
		os.Exit(util.ErrorExitCode)
	}

	controller := &upgradeController{
		installParam: installParam,
		logPath:      installParam.LogDir,
		logBackPath:  installParam.LogBackupDir,
	}

	hwlog.RunLog.Info("-------------------start to upgrade MEF-Center steps in new package-------------------")
	if err = controller.doUpgrade(); err != nil {
		hwlog.RunLog.Error("-------------------upgrade MEF-Center steps in new package failed-------------------")
		os.Exit(util.ErrorExitCode)
	}
	hwlog.RunLog.Info("-------------------end to upgrade MEF-Center steps in new package-------------------")
}

func initLog(installParam *util.InstallParamJsonTemplate) error {
	logDirPath := installParam.LogDir
	logBackupDirPath := installParam.LogBackupDir
	logPathMgr := util.InitLogDirPathMgr(logDirPath, logBackupDirPath)
	logPath, err := utils.CheckPath(logPathMgr.GetInstallLogPath())
	if err != nil {
		return fmt.Errorf("check log path %s failed:%s", logPath, err.Error())
	}
	logBackupPath, err := utils.CheckPath(logPathMgr.GetInstallLogBackupPath())
	if err != nil {
		return fmt.Errorf("check log backup path %s failed:%s", logBackupPath, err.Error())
	}

	if err = util.InitLogPath(logPath, logBackupPath); err != nil {
		return fmt.Errorf("init log path %s failed:%s", logPath, err.Error())
	}
	return nil
}

func (uc upgradeController) doUpgrade() error {
	installedComponents := util.GetCompulsorySlice()

	controlMgr := control.GetUpgradePostMgr(installedComponents, uc.installParam.InstallDir, uc.logPath, uc.logBackPath)

	if err := controlMgr.DoUpgrade(); err != nil {
		return err
	}
	return nil
}
