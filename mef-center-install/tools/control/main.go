// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package main manages MEF cloud start, stop and restart
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindxedge/base/mef-center-install/pkg/control"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

var (
	componentType string
)

const (
	startFlag   = "start"
	stopFlag    = "stop"
	restartFlag = "restart"
)

func init() {
	flag.StringVar(&componentType, startFlag, "all", "start a component, default all components")
	flag.StringVar(&componentType, stopFlag, "all", "stop a component, default all components")
	flag.StringVar(&componentType, restartFlag, "all", "restart a component, default all components")
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
	flags := [3]string{startFlag, stopFlag, restartFlag}
	for _, s := range flags {
		if isFlagSet(s) {
			return s
		}
	}
	return ""
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

func doControl(operate string, installParam *util.InstallParamJsonTemplate) error {
	pathMgr := util.InitInstallDirPathMgr(installParam.InstallDir)

	installedComponents := installParam.Components
	if err := checkComponent(installedComponents); err != nil {
		return err
	}

	controlMgr := control.InitSftControlMgr(componentType, operate, installedComponents, pathMgr)
	if err := controlMgr.DoControl(); err != nil {
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
		os.Exit(1)
	}

	logDirPath := installParam.LogDir
	logPathMgr := util.InitLogDirPathMgr(logDirPath)
	logPath := logPathMgr.GetInstallLogPath()
	if logPath, err = utils.CheckPath(logPath); err != nil {
		fmt.Printf("check log path %s failed:%s\n", logPath, err.Error())
		os.Exit(1)
	}

	if err = util.InitLogPath(logPath); err != nil {
		fmt.Printf("init log path %s failed:%s\n", logPath, err.Error())
		os.Exit(1)
	}

	hwlog.RunLog.Errorf("start to %s %s component", operate, componentType)
	hwlog.OpLog.Errorf("start to %s %s component", operate, componentType)
	if err = doControl(operate, installParam); err != nil {
		hwlog.RunLog.Errorf("%s %s component failed", operate, componentType)
		hwlog.OpLog.Errorf("%s %s component failed", operate, componentType)
		os.Exit(1)
	}
	hwlog.RunLog.Infof("%s %s component successful", operate, componentType)
	hwlog.OpLog.Infof("%s %s component successful", operate, componentType)
	fmt.Printf("%s %s component successful\n", operate, componentType)
}
