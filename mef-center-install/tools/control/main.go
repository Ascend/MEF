// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package main manages MEF cloud start, stop and restart
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"

	"huawei.com/mindxedge/base/mef-center-install/pkg/control"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

type controller interface {
	doControl() error
	setInstallParam(installParam *util.InstallParamJsonTemplate)
	bindFlag() bool
	printExecutingLog(ip, user string)
	printFailedLog(ip, user string)
	printSuccessLog(ip, user string)
}

type operateController struct {
	operate      string
	installParam *util.InstallParamJsonTemplate
}

type uninstallController struct {
	installParam *util.InstallParamJsonTemplate
}

type upgradeController struct {
	installParam *util.InstallParamJsonTemplate
}

type logExportController struct {
	module       string
	edgeNodes    string
	installParam *util.InstallParamJsonTemplate
}

var (
	// BuildName the program name
	BuildName string
	// BuildVersion the program version
	BuildVersion string

	componentType string
	version       bool
	zipPath       string
	help          bool
	curController controller

	allowedModule = []string{util.EdgeManagerName, util.NginxManagerName, util.CertManagerName}
)

const (
	componentFlag = "component"
	pathFlag      = "pkg_path"
	moduleFlag    = "module"
	nodesFlag     = "nodes"
)

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

func (oc *operateController) setInstallParam(installParam *util.InstallParamJsonTemplate) {
	oc.installParam = installParam
}

func (oc *operateController) bindFlag() bool {
	flag.StringVar(&componentType, componentFlag, "all", "start、stop、restart a component, default all components")
	return true
}

func (oc *operateController) printExecutingLog(ip, user string) {
	fmt.Printf("start to %s %s component\n", oc.operate, componentType)
	hwlog.RunLog.Infof("-------------------start to %s %s component-------------------", oc.operate, componentType)
	hwlog.OpLog.Infof("%s: %s, start to %s %s component", ip, user, oc.operate, componentType)
}

func (oc *operateController) printFailedLog(ip, user string) {
	fmt.Printf("%s %s component failed\n", oc.operate, componentType)
	hwlog.RunLog.Errorf("-------------------%s %s component failed-------------------", oc.operate, componentType)
	hwlog.OpLog.Errorf("%s: %s, %s %s component failed", ip, user, oc.operate, componentType)
}

func (oc *operateController) printSuccessLog(ip, user string) {
	fmt.Printf("%s %s component successful\n", oc.operate, componentType)
	hwlog.RunLog.Infof("-------------------%s %s component successful-------------------", oc.operate, componentType)
	hwlog.OpLog.Infof("%s: %s, %s %s component successful", ip, user, oc.operate, componentType)
}

func (uc *uninstallController) doControl() error {
	installedComponents := uc.installParam.Components

	controlMgr := control.GetSftUninstallMgrIns(installedComponents, uc.installParam.InstallDir)
	if err := controlMgr.DoUninstall(); err != nil {
		return err
	}
	return nil
}

func (uc *uninstallController) setInstallParam(installParam *util.InstallParamJsonTemplate) {
	uc.installParam = installParam
}

func (uc *uninstallController) bindFlag() bool {
	return false
}

func (uc *uninstallController) printExecutingLog(ip, user string) {
	hwlog.RunLog.Info("-------------------start to uninstall MEF-Center-------------------")
	hwlog.OpLog.Infof("%s: %s, start to uninstall MEF-Center", ip, user)
	fmt.Println("start to uninstall MEF-Center")
}

func (uc *uninstallController) printFailedLog(ip, user string) {
	hwlog.RunLog.Error("-------------------uninstall MEF-Center failed-------------------")
	hwlog.OpLog.Errorf("%s: %s, uninstall MEF-Center failed", ip, user)
	fmt.Println("uninstall MEF-Center failed")
}

func (uc *uninstallController) printSuccessLog(ip, user string) {
	hwlog.RunLog.Info("-------------------uninstall MEF-Center successful-------------------")
	hwlog.OpLog.Infof("%s: %s, uninstall MEF-Center successful", ip, user)
	fmt.Println("uninstall MEF-Center successful")
}

func (uc *upgradeController) doControl() error {
	installedComponents := uc.installParam.Components

	if err := uc.checkZipPath(); err != nil {
		hwlog.RunLog.Errorf("check zip path failed: %s", err.Error())
		return err
	}
	controlMgr := control.GetUpgradePreMgr(zipPath, installedComponents, uc.installParam.InstallDir)

	if err := controlMgr.DoUpgrade(); err != nil {
		return err
	}
	return nil
}

func (uc *upgradeController) checkZipPath() error {
	const zipSizeMul int64 = 100

	pathMgr := util.InitInstallDirPathMgr(uc.installParam.InstallDir)
	unpackPath := pathMgr.WorkPathMgr.GetVarDirPath()
	if filepath.Dir(zipPath) == unpackPath {
		fmt.Println("zip path cannot be inside the unpack path")
		hwlog.RunLog.Errorf("zip path cannot be the unpack dir:%s", unpackPath)
		return errors.New("zip path cannot be the unpack dir")
	}

	if zipPath == "" || !utils.IsExist(zipPath) {
		fmt.Println("zip path does not exist")
		return errors.New("zip path does not exist")
	}

	ret, err := common.IfAbsPath(zipPath)
	if err != nil {
		fmt.Println("get zip path's abs path failed")
		return fmt.Errorf("get path [%s]'s abs path failed: %s", zipPath, err.Error())
	}
	if !ret {
		fmt.Println("zip path is not an absolute path")
		return fmt.Errorf("zip path is not abs path")
	}

	if _, err = utils.RealFileChecker(zipPath, true, false, zipSizeMul); err != nil {
		fmt.Printf("check zip path failed: %s\n", err.Error())
		return fmt.Errorf("zip path check failed: %s", err.Error())
	}

	return nil
}

func (uc *upgradeController) setInstallParam(installParam *util.InstallParamJsonTemplate) {
	uc.installParam = installParam
}

func (uc *upgradeController) bindFlag() bool {
	flag.StringVar(&zipPath, pathFlag, "", "the path of the zip file to upgrade MEF Center")
	return true
}

func (uc *upgradeController) printExecutingLog(ip, user string) {
	hwlog.RunLog.Info("-------------------start to upgrade MEF-Center-------------------")
	hwlog.OpLog.Infof("%s: %s, start to upgrade MEF-Center", ip, user)
	fmt.Println(" start to upgrade MEF-Center")
}

func (uc *upgradeController) printFailedLog(ip, user string) {
	hwlog.RunLog.Error("-------------------upgrade MEF-Center failed-------------------")
	hwlog.OpLog.Errorf("%s: %s, upgrade MEF-Center failed", ip, user)
	fmt.Println("upgrade MEF-Center failed")
}

func (uc *upgradeController) printSuccessLog(ip, user string) {
	hwlog.RunLog.Info("-------------------upgrade MEF-Center successful-------------------")
	hwlog.OpLog.Infof("%s: %s, upgrade MEF-Center successful", ip, user)
	fmt.Println("upgrade MEF-Center successful")
}

func (lec *logExportController) doControl() error {
	logDirPathMgr := util.InitLogDirPathMgr(lec.installParam.LogDir, lec.installParam.LogBackupDir)
	installDirPathMgr := util.InitInstallDirPathMgr(lec.installParam.InstallDir)
	controlMgr := control.GetLogExportMgrIns(
		lec.module, strings.Split(lec.edgeNodes, ","), logDirPathMgr, installDirPathMgr)
	if err := controlMgr.DoExport(); err != nil {
		return err
	}
	return nil
}

func (lec *logExportController) setInstallParam(installParam *util.InstallParamJsonTemplate) {
	lec.installParam = installParam
}

func (lec *logExportController) bindFlag() bool {
	flag.StringVar(&lec.module, moduleFlag, "", "the module to export logs")
	flag.StringVar(&lec.edgeNodes, nodesFlag, "", "the nodes to export logs")
	return true
}

func (lec *logExportController) printExecutingLog(ip, user string) {
	hwlog.RunLog.Info("-------------------start to export logs-------------------")
	hwlog.OpLog.Infof("%s: %s, start to export logs", ip, user)
	fmt.Println("start to export logs")
}

func (lec *logExportController) printFailedLog(ip, user string) {
	hwlog.RunLog.Error("-------------------export logs failed-------------------")
	hwlog.OpLog.Errorf("%s: %s, export logs failed", ip, user)
	fmt.Println("export logs failed")
}

func (lec *logExportController) printSuccessLog(ip, user string) {
	hwlog.RunLog.Info("-------------------export logs successful-------------------")
	hwlog.OpLog.Infof("%s: %s, export logs successful", ip, user)
	fmt.Println("export logs successful")
}

func dealArgs() bool {
	flag.Usage = printUseHelp
	if len(os.Args) == util.NoArgCount {
		printUseHelp()
		return false
	}
	if os.Args[util.CtlArgIndex][0] == '-' {
		return dealControlFlag()
	}
	return dealCmdFlag()
}

func dealControlFlag() bool {
	flag.BoolVar(&version, "version", false, "")
	flag.BoolVar(&help, "h", false, "")
	flag.BoolVar(&help, "help", false, "")
	flag.Parse()
	if help {
		printUsage()
		return false
	}
	if version {
		printVersion()
		return false
	}
	printUseHelp()
	return false
}

func dealCmdFlag() bool {
	operate := os.Args[util.CmdIndex]
	optMap := getOperateMap(operate)
	operator, ok := optMap[operate]
	if !ok {
		fmt.Println("the parameter is invalid")
		printUseHelp()
		return false
	}

	curController = operator
	if !operator.bindFlag() {
		return true
	}

	flag.Usage = flag.PrintDefaults
	if err := flag.CommandLine.Parse(os.Args[util.CmdArgIndex:]); err != nil {
		fmt.Printf("parse cmd args failed,error:%v\n", err)
		return false
	}
	return true
}

func printUseHelp() {
	fmt.Println("use '-help' for help information")
}

func printVersion() {
	fmt.Printf("%s version: %s\n", BuildName, BuildVersion)
}

func printUsage() {
	printVersion()
	fmt.Printf(`Usage: [OPTIONS...] COMMAND

Options:
	-help		Print help information
	-version	Print version information

Commands:
	start      -- start all or a component
	stop       -- stop all or a component
	restart    -- restart all or a component
	uninstall  -- uninstall MEF Center
	upgrade    -- upgrade MEF Center
	exportlog  -- export MEF logs
`)
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

	if !dealArgs() {
		return
	}
	fmt.Println("init log success")
	user, ip, err := util.GetLoginUserAndIP()
	if err != nil {
		hwlog.RunLog.Errorf("get current user or ip info failed: %s", err.Error())
		os.Exit(util.ErrorExitCode)
	}

	curController.setInstallParam(installParam)
	curController.printExecutingLog(ip, user)
	if err = curController.doControl(); err != nil {
		curController.printFailedLog(ip, user)
		os.Exit(util.ErrorExitCode)
	}
	curController.printSuccessLog(ip, user)
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

func getOperateMap(operate string) map[string]controller {
	return map[string]controller{
		util.StartOperateFlag:   &operateController{operate: operate},
		util.StopOperateFlag:    &operateController{operate: operate},
		util.RestartOperateFlag: &operateController{operate: operate},
		util.UninstallFlag:      &uninstallController{},
		util.UpgradeFlag:        &upgradeController{},
		util.LogExportFlag:      &logExportController{},
	}
}
