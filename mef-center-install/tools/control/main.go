// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package main manages MEF cloud start, stop and restart
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/terminal"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindxedge/base/mef-center-install/pkg/control"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

type controller interface {
	doControl() error
}

type operateController struct {
	operate      string
	installParam *util.InstallParamJsonTemplate
}

type uninstallController struct {
	installParam *util.InstallParamJsonTemplate
}

type upgradeController struct {
	zipPath      string
	installParam *util.InstallParamJsonTemplate
}

var (
	// BuildName the program name
	BuildName string
	// BuildVersion the program version
	BuildVersion string

	componentType string
	version       bool
	operateType   string
	zipPath       string

	allowedModule = []string{util.EdgeManagerName, util.NginxManagerName, util.CertManagerName}
)

const (
	startFlag   = "start"
	stopFlag    = "stop"
	restartFlag = "restart"
	operateFlag = "operate"
	pathFlag    = "zipPath"
)

func init() {
	flag.StringVar(&componentType, startFlag, "all", "start a component, default all components")
	flag.StringVar(&componentType, stopFlag, "all", "stop a component, default all components")
	flag.StringVar(&componentType, restartFlag, "all", "restart a component, default all components")
	flag.StringVar(&operateType, operateFlag, "other", "to illustrate the operate type: control, uninstall or upgrade")
	flag.StringVar(&zipPath, pathFlag, "", "the path of the zip file to upgrade MEF Center")
	flag.BoolVar(&version, util.VersionFlag, false, "Output the program version")
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

func checkFlag() string {
	// the first operate type will be performed
	flags := [util.RunFlagCount]string{startFlag, stopFlag, restartFlag}
	for _, s := range flags {
		if isFlagSet(s) {
			return s
		}
	}
	return operateType
}

func checkComponent(installedComponents []string) error {
	var validType bool
	if componentType == "all" {
		return nil
	}

	for _, component := range allowedModule {
		if component == componentType {
			validType = true
			break
		}
	}

	if !validType {
		fmt.Println("unsupported component")
		hwlog.RunLog.Errorf("unsupported component")
		return errors.New("unsupported component")
	}

	for _, component := range installedComponents {
		if component == componentType {
			return nil
		}
	}

	hwlog.RunLog.Errorf("the component %s is not installed yet", componentType)
	return errors.New("the target component is not installed")
}

func (oc *operateController) doControl() error {
	installedComponents := oc.installParam.Components
	if err := checkComponent(installedComponents); err != nil {
		return err
	}

	controlMgr := control.InitSftOperateMgr(componentType, oc.operate,
		installedComponents, util.InitInstallDirPathMgr(oc.installParam.InstallDir),
		util.InitLogDirPathMgr(oc.installParam.LogDir, oc.installParam.LogBackupDir))
	if err := controlMgr.DoOperate(); err != nil {
		return err
	}
	return nil
}

func (oc *uninstallController) doControl() error {
	installedComponents := oc.installParam.Components

	controlMgr := control.GetSftUninstallMgrIns(installedComponents, oc.installParam.InstallDir)
	if err := controlMgr.DoUninstall(); err != nil {
		return err
	}
	return nil
}

func (uc *upgradeController) doControl() error {
	installedComponents := uc.installParam.Components
	var zipSizeMul int64 = 100

	pathMgr := util.InitInstallDirPathMgr(uc.installParam.InstallDir)
	unpackPath := pathMgr.WorkPathMgr.GetRelativeVarDirPath()
	if filepath.Dir(uc.zipPath) == unpackPath {
		hwlog.RunLog.Errorf("zipDir cannot be the unpack dir:%s", unpackPath)
		return errors.New("zipDir cannot be the unpack dir")
	}

	if _, err := utils.RealFileChecker(uc.zipPath, true, false, zipSizeMul); err != nil {
		hwlog.RunLog.Errorf("zipPath check failed: %s", err)
		return errors.New("zipPath check failed")
	}

	controlMgr := control.GetUpgradePreMgr(uc.zipPath, installedComponents, uc.installParam.InstallDir)

	if err := controlMgr.DoUpgrade(); err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Parse()

	if version {
		fmt.Printf("%s version: %s\n", BuildName, BuildVersion)
		os.Exit(util.VersionExitCode)
	}

	operate := checkFlag()
	installParam, err := util.GetInstallInfo()
	if err != nil {
		fmt.Printf("get info from install-param.json failed:%s\n", err.Error())
		os.Exit(util.ErrorExitCode)
	}

	if err = initLog(installParam); err != nil {
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

	operateMap := getOperateMap(installParam, operate)
	controllerIns := operateMap[operateType]
	if controllerIns == nil {
		hwlog.RunLog.Error("get controller failed")
		hwlog.OpLog.Errorf("%s: %s,  unsupported operate type", ip, user)
		os.Exit(util.ErrorExitCode)
	}

	hwlog.RunLog.Infof("-------------------start to %s %s component-------------------", operate, componentType)
	hwlog.OpLog.Infof("%s: %s, start to %s %s component", ip, user, operate, componentType)
	if err = controllerIns.doControl(); err != nil {
		hwlog.RunLog.Errorf("-------------------%s %s component failed-------------------", operate, componentType)
		hwlog.OpLog.Errorf("%s: %s, %s %s component failed", ip, user, operate, componentType)
		os.Exit(util.ErrorExitCode)
	}
	hwlog.RunLog.Infof("-------------------%s %s component successful-------------------", operate, componentType)
	hwlog.OpLog.Infof("%s: %s, %s %s component successful", ip, user, operate, componentType)
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

	if err := util.InitLogPath(logPath, logBackupPath); err != nil {
		return fmt.Errorf("init log path %s failed:%s", logPath, err.Error())
	}
	return nil
}

func getOperateMap(installParam *util.InstallParamJsonTemplate, operate string) map[string]controller {
	return map[string]controller{
		util.OperateFlag:   &operateController{operate: operate, installParam: installParam},
		util.UninstallFlag: &uninstallController{installParam: installParam},
		util.UpgradeFlag: &upgradeController{
			installParam: installParam,
			zipPath:      zipPath,
		},
	}
}
