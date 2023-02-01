// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package main manages MEF cloud start, stop and restart
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

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

var (
	componentType string
	operateType   string
)

const (
	startFlag   = "start"
	stopFlag    = "stop"
	restartFlag = "restart"
	operateFlag = "operate"
)

func init() {
	flag.StringVar(&componentType, startFlag, "all", "start a component, default all components")
	flag.StringVar(&componentType, stopFlag, "all", "stop a component, default all components")
	flag.StringVar(&componentType, restartFlag, "all", "restart a component, default all components")
	flag.StringVar(&operateType, operateFlag, "",
		"to illustrate the operate type: control, uninstall or upgrade")
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
	if componentType == "all" {
		return nil
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

	controlMgr := control.InitSftOperateMgr(componentType, oc.operate, installedComponents, oc.installParam.InstallDir)
	if err := controlMgr.DoOperate(); err != nil {
		return err
	}
	return nil
}

func (oc *uninstallController) doControl() error {
	installedComponents := oc.installParam.Components
	if err := checkComponent(installedComponents); err != nil {
		return err
	}

	controlMgr := control.GetSftUninstallMgrIns(installedComponents, oc.installParam.InstallDir)
	if err := controlMgr.DoUninstall(); err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Parse()
	fmt.Println("in main control")
	operate := checkFlag()

	installParam, err := util.GetInstallInfo()
	if err != nil {
		fmt.Printf("get info from install-param.json failed:%s\n", err.Error())
		os.Exit(util.ErrorExitCode)
	}

	logDirPath := installParam.LogDir
	logPathMgr := util.InitLogDirPathMgr(logDirPath)
	logPath := logPathMgr.GetInstallLogPath()
	if logPath, err = utils.CheckPath(logPath); err != nil {
		fmt.Printf("check log path %s failed:%s\n", logPath, err.Error())
		os.Exit(util.ErrorExitCode)
	}

	if err = util.InitLogPath(logPath); err != nil {
		fmt.Printf("init log path %s failed:%s\n", logPath, err.Error())
		os.Exit(util.ErrorExitCode)
	}
	fmt.Println("init log success")

	user, ip, err := terminal.GetLoginUserAndIP()
	if err != nil {
		hwlog.RunLog.Errorf("get current user or ip info failed: %s", err.Error())
		hwlog.OpLog.Error("install MEF Center failed: cannot get local user or ip")
		os.Exit(util.ErrorExitCode)
	}

	operateMap := map[string]controller{
		util.OperateFlag:   &operateController{operate: operate, installParam: installParam},
		util.UninstallFlag: &uninstallController{installParam: installParam},
	}
	controllerIns := operateMap[operateType]
	if controllerIns == nil {
		hwlog.RunLog.Error("get controller failed")
		hwlog.OpLog.Errorf("%s: %s,  unsupported operate type", ip, user)
		os.Exit(util.ErrorExitCode)
	}

	hwlog.RunLog.Infof("start to %s %s component", operate, componentType)
	hwlog.OpLog.Infof("%s: %s, start to %s %s component", ip, user, operate, componentType)
	if err = controllerIns.doControl(); err != nil {
		hwlog.RunLog.Errorf("%s %s component failed", operate, componentType)
		hwlog.OpLog.Errorf("%s: %s, %s %s component failed", ip, user, operate, componentType)
		os.Exit(1)
	}
	hwlog.RunLog.Infof("%s %s component successful", operate, componentType)
	hwlog.OpLog.Infof("%s: %s, %s %s component successful", ip, user, operate, componentType)
}
