// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package main manages MEF cloud start, stop and restart
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/mef-center-install/pkg/control"
	"huawei.com/mindxedge/base/mef-center-install/pkg/control/kmcupdate"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

type controller interface {
	doControl() error
	setInstallParam(installParam *util.InstallParamJsonTemplate)
	bindFlag() bool
	printExecutingLog(ip, user string)
	printFailedLog(ip, user string)
	printSuccessLog(ip, user string)
	getName() string
}

type operateController struct {
	operate      string
	installParam *util.InstallParamJsonTemplate
}

type uninstallController struct {
	operate      string
	installParam *util.InstallParamJsonTemplate
}

var (
	// BuildName the program name
	BuildName string
	// BuildVersion the program version
	BuildVersion string

	componentType string
	version       bool
	help          bool
	curController controller
)

const (
	componentFlag = "component"
	tarPathFlag   = "file"
	cmsPathFlag   = "cms"
	crlPathFlag   = "crl"
)

func checkComponent(components []string) error {
	if componentType == "all" {
		return nil
	}
	allowedModule := append(util.GetCompulsorySlice(), util.OptionalComponent()...)

	if err := util.CheckParamOption(allowedModule, componentType); err != nil {
		fmt.Println("unsupported component")
		hwlog.RunLog.Errorf("unsupported component")
		return errors.New("unsupported component")
	}

	installInfo, err := util.GetInstallInfo()
	if err != nil {
		return err
	}
	installedComponents := append(components, installInfo.OptionComponent...)

	if err := util.CheckParamOption(installedComponents, componentType); err != nil {
		fmt.Printf("the component %s is not installed yet\n", componentType)
		hwlog.RunLog.Errorf("the component %s is not installed yet", componentType)
		return errors.New("the target component is not installed")
	}

	return nil

}

func (oc *operateController) doControl() error {
	components := util.GetCompulsorySlice()
	if err := checkComponent(components); err != nil {
		return err
	}

	installPathMgr, err := util.InitInstallDirPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("init install path mgr failed: %v", err)
		return errors.New("init install path mgr failed")
	}
	controlMgr := control.InitSftOperateMgr(componentType, oc.operate, components, installPathMgr,
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
	hwlog.OpLog.Infof("[%s@%s] start to %s %s component", user, ip, oc.operate, componentType)
}

func (oc *operateController) printFailedLog(ip, user string) {
	fmt.Printf("%s %s component failed\n", oc.operate, componentType)
	hwlog.RunLog.Errorf("-------------------%s %s component failed-------------------", oc.operate, componentType)
	hwlog.OpLog.Errorf("[%s@%s] %s %s component failed", user, ip, oc.operate, componentType)
}

func (oc *operateController) printSuccessLog(ip, user string) {
	fmt.Printf("%s %s component successful\n", oc.operate, componentType)
	hwlog.RunLog.Infof("-------------------%s %s component successful-------------------", oc.operate, componentType)
	hwlog.OpLog.Infof("[%s@%s] %s %s component successful", user, ip, oc.operate, componentType)
}

func (oc *operateController) getName() string {
	return oc.operate
}

func (uc *uninstallController) doControl() error {
	installedComponents := util.GetCompulsorySlice()

	installPathMgr, err := util.InitInstallDirPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("init install path mgr failed: %v", err)
		return errors.New("init install path mgr failed")
	}
	controlMgr := control.GetSftUninstallMgrIns(installedComponents, installPathMgr)
	if err := controlMgr.DoUninstall(); err != nil {
		hwlog.RunLog.Errorf("uninstall failed, error: %v", err)
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
	hwlog.OpLog.Infof("[%s@%s] start to uninstall MEF-Center", user, ip)
	fmt.Println("start to uninstall MEF-Center")
}

func (uc *uninstallController) printFailedLog(ip, user string) {
	hwlog.RunLog.Error("-------------------uninstall MEF-Center failed-------------------")
	hwlog.OpLog.Errorf("[%s@%s] uninstall MEF-Center failed", user, ip)
	fmt.Println("uninstall MEF-Center failed")
}

func (uc *uninstallController) printSuccessLog(ip, user string) {
	hwlog.RunLog.Info("-------------------uninstall MEF-Center successful-------------------")
	hwlog.OpLog.Infof("[%s@%s] uninstall MEF-Center successful", user, ip)
	fmt.Println("uninstall MEF-Center successful")
}

func (uc *uninstallController) getName() string {
	return uc.operate
}

type upgradeController struct {
	operate      string
	installParam *util.InstallParamJsonTemplate
	tarPath      string
	cmsPath      string
	crlPath      string
}

func (uc *upgradeController) doControl() error {
	installedComponents := util.GetCompulsorySlice()

	controlMgr, err := control.GetUpgradePreMgr(uc.tarPath, uc.cmsPath, uc.crlPath, installedComponents)
	if err != nil {
		hwlog.RunLog.Errorf("get upgrade pre mgr failed: %v", err)
		return err
	}

	if err := controlMgr.DoUpgrade(); err != nil {
		hwlog.RunLog.Errorf("upgrade failed, error: %v", err)
		return err
	}
	return nil
}

func (uc *upgradeController) setInstallParam(installParam *util.InstallParamJsonTemplate) {
	uc.installParam = installParam
}

func (uc *upgradeController) bindFlag() bool {
	flag.StringVar(&(uc.tarPath), tarPathFlag, "", "path of the software upgrade tar.gz file")
	flag.StringVar(&(uc.cmsPath), cmsPathFlag, "", "path of the software upgrade tar.gz.cms file")
	flag.StringVar(&(uc.crlPath), crlPathFlag, "", "path of the software upgrade tar.gz.crl file")
	utils.MarkFlagRequired(tarPathFlag)
	utils.MarkFlagRequired(cmsPathFlag)
	utils.MarkFlagRequired(crlPathFlag)
	return true
}

func (uc *upgradeController) printExecutingLog(ip, user string) {
	hwlog.RunLog.Info("-------------------start to upgrade MEF-Center-------------------")
	hwlog.OpLog.Infof("[%s@%s] start to upgrade MEF-Center", user, ip)
	fmt.Println(" start to upgrade MEF-Center")
}

func (uc *upgradeController) printFailedLog(ip, user string) {
	hwlog.RunLog.Error("-------------------upgrade MEF-Center failed-------------------")
	hwlog.OpLog.Errorf("[%s@%s] upgrade MEF-Center failed", user, ip)
	fmt.Println("upgrade MEF-Center failed")
}

func (uc *upgradeController) printSuccessLog(ip, user string) {
	hwlog.RunLog.Info("-------------------upgrade MEF-Center successful-------------------")
	hwlog.OpLog.Infof("[%s@%s] upgrade MEF-Center successful", user, ip)
	fmt.Println("upgrade MEF-Center successful")
}

func (uc *upgradeController) getName() string {
	return uc.operate
}

type exchangeCertsController struct {
	operate      string
	installParam *util.InstallParamJsonTemplate
	importPath   string
	exportPath   string
	component    string
}

const (
	importPathFlag = "import_path"
	exportPathFlag = "export_path"
)

func (ecc *exchangeCertsController) bindFlag() bool {
	flag.StringVar(&(ecc.importPath), importPathFlag, "", "path that saves ca cert to import")
	flag.StringVar(&(ecc.exportPath), exportPathFlag, "", "path to export MEF ca cert")
	utils.MarkFlagRequired(importPathFlag)
	utils.MarkFlagRequired(exportPathFlag)
	return true
}

func (ecc *exchangeCertsController) setInstallParam(installParam *util.InstallParamJsonTemplate) {
	ecc.installParam = installParam
}

func (ecc *exchangeCertsController) doControl() error {
	pathMgr, err := util.InitInstallDirPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("init install path mgr failed: %v", err)
		return errors.New("init install path mgr failed")
	}
	exchangeFlow, err := control.NewExchangeCaFlow(ecc.importPath, ecc.exportPath, util.NginxManagerName, pathMgr)
	if err != nil {
		return err
	}
	if err = exchangeFlow.DoExchange(); err != nil {
		hwlog.RunLog.Errorf("execute exchange flow failed: %s", err.Error())
		return err
	}

	return nil
}

func (ecc *exchangeCertsController) printExecutingLog(ip, user string) {
	hwlog.RunLog.Info("-------------------start to exchange certs-------------------")
	hwlog.OpLog.Infof("[%s@%s] start to exchange certs", user, ip)
	fmt.Println("start to exchange certs")
}

func (ecc *exchangeCertsController) printSuccessLog(ip, user string) {
	hwlog.RunLog.Info("-------------------exchange certs successful-------------------")
	hwlog.OpLog.Infof("[%s@%s] exchange certs successful", user, ip)
	fmt.Println("exchange certs successful")
}

func (ecc *exchangeCertsController) printFailedLog(ip, user string) {
	hwlog.RunLog.Error("-------------------exchange certs failed-------------------")
	hwlog.OpLog.Errorf("[%s@%s] exchange certs failed", user, ip)
	fmt.Println("exchange certs failed, for more information please look up mef install log files")
}

func (ecc *exchangeCertsController) getName() string {
	return ecc.operate
}

type updateKmcController struct {
	operate      string
	installParam *util.InstallParamJsonTemplate
}

func (ukc *updateKmcController) bindFlag() bool {
	return false
}

func (ukc *updateKmcController) setInstallParam(installParam *util.InstallParamJsonTemplate) {
	ukc.installParam = installParam
}

func (ukc *updateKmcController) doControl() error {
	pathMgr, err := util.InitInstallDirPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("init install path mgr failed: %v", err)
		return errors.New("init install path mgr failed")
	}

	updateFlow := kmcupdate.NewUpdateKmcFlow(pathMgr)
	if err := updateFlow.RunFlow(); err != nil {
		hwlog.RunLog.Errorf("execute update kmc flow failed: %s", err.Error())
		return err
	}

	return nil
}

func (ukc *updateKmcController) printExecutingLog(ip, user string) {
	hwlog.RunLog.Info("-------------------start to update kmc keys-------------------")
	hwlog.OpLog.Infof("[%s@%s] start to update kmc keys", user, ip)
	fmt.Println(" start to update kmc keys")
}

func (ukc *updateKmcController) printSuccessLog(ip, user string) {
	hwlog.RunLog.Info("-------------------update kmc keys successful-------------------")
	hwlog.OpLog.Infof("[%s@%s] update kmc keys successful", user, ip)
	fmt.Println("update kmc keys successful")
}

func (ukc *updateKmcController) printFailedLog(ip, user string) {
	hwlog.RunLog.Error("-------------------update kmc keys failed-------------------")
	hwlog.OpLog.Errorf("[%s@%s] update kmc keys failed", user, ip)
	fmt.Println("update kmc keys failed")
}

func (ukc *updateKmcController) getName() string {
	return ukc.operate
}

type manageThirdComponent struct {
	installParam *util.InstallParamJsonTemplate
	component    string
	operate      string
	lockOperate  string
	control.SubParam
}

const (
	operateFlag            = "operate"
	installPackagePathFlag = "install_tar_file"
	installCmsPathFlag     = "install_cms_file"
	installCrlPathFlag     = "install_crl_file"
)

func (mtc *manageThirdComponent) bindFlag() bool {
	flag.StringVar(&(mtc.component), componentFlag, "", "component name, only support [ics-manager]")
	flag.StringVar(&(mtc.operate), operateFlag, "", "manage third component operate, only support [install, uninstall]")
	flag.StringVar(&(mtc.InstallPackagePath), installPackagePathFlag, "",
		"install package tar.gz file path, install operate necessary parameter")
	flag.StringVar(&(mtc.InstallCmsPath), installCmsPathFlag, "",
		"install package cms file path, install operate necessary parameter")
	flag.StringVar(&(mtc.InstallCrlPath), installCrlPathFlag, "",
		"install package crl file path, install operate necessary parameter")
	utils.MarkFlagRequired(componentFlag)
	utils.MarkFlagRequired(operateFlag)
	return true
}

func (mtc *manageThirdComponent) setInstallParam(installParam *util.InstallParamJsonTemplate) {
	mtc.installParam = installParam
}

func (mtc *manageThirdComponent) doControl() error {
	pathMgr, err := util.InitInstallDirPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("init install path mgr failed: %v", err)
		return errors.New("init install path mgr failed")
	}

	exchangeFlow := control.NewThirdComponentManageFlow(mtc.component, mtc.operate, mtc.SubParam, pathMgr)
	if err = exchangeFlow.DoManage(); err != nil {
		hwlog.RunLog.Errorf("%s %s failed: %s", mtc.operate, mtc.component, err.Error())
		return err
	}

	return nil
}

func (mtc *manageThirdComponent) printExecutingLog(ip, user string) {
	hwlog.RunLog.Info("-------------------start to manage third component-------------------")
	hwlog.OpLog.Infof("[%s@%s] start to manage third component", user, ip)
	fmt.Println("start to manage third component")
}

func (mtc *manageThirdComponent) printSuccessLog(ip, user string) {
	hwlog.RunLog.Info("-------------------manage third component successful-------------------")
	hwlog.OpLog.Infof("[%s@%s] manage third component successful", user, ip)
	fmt.Println("manage third component successful")
}

func (mtc *manageThirdComponent) printFailedLog(ip, user string) {
	hwlog.RunLog.Error("-------------------manage third component failed-------------------")
	hwlog.OpLog.Errorf("[%s@%s] manage third component failed", user, ip)
	fmt.Println("manage third component failed")
}

func (mtc *manageThirdComponent) getName() string {
	return mtc.lockOperate
}

func dealArgs() bool {
	flag.Usage = printUseHelp
	if len(os.Args) <= util.NoArgCount {
		printUseHelp()
		return false
	}

	if len(os.Args[util.CtlArgIndex]) == 0 {
		fmt.Println("the required parameter is missing")
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
	if utils.IsRequiredFlagNotFound() {
		fmt.Println("the required parameter is missing")
		flag.PrintDefaults()
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
	fmt.Printf(`Usage: [OPTIONS...] COMMAND

Options:
	-help		Print help information
	-version	Print version information

Commands:
	start       	-- start all or a component
	stop        	-- stop all or a component
	restart     	-- restart all or a component
	uninstall   	-- uninstall MEF Center
	upgrade     	-- upgrade MEF Center
	exchangeca  	-- exchange root ca with MEF Center
	updatekmc   	-- update kmc keys
	importcrl   	-- improt crl from the Northbound ca
	alarmconfig 	-- update alarm used configuration 
	getalarmconfig  -- get alarm used configuration
	managecomponent -- manage third component
`)
}

const retryTimes = 3

var (
	installParam *util.InstallParamJsonTemplate
	err          error
)

func initInstallParam() bool {
	var readSuccess bool
	for i := 1; i <= retryTimes; i++ {
		installParam, err = util.GetInstallInfo()
		if err != nil {
			fmt.Printf("get info from install-param.json failed:%s\n", err.Error())
			continue
		}
		readSuccess = true
		break
	}

	return readSuccess
}

func main() {
	errorCode := handle()
	os.Exit(errorCode)
}

func handle() int {
	if !initInstallParam() {
		return util.ErrorExitCode
	}
	if err = initLog(installParam); err != nil {
		fmt.Println(err.Error())
		return util.ErrorExitCode
	}

	if !dealArgs() {
		return util.ErrorExitCode
	}
	fmt.Println("init log success")
	user, ip, err := envutils.GetUserAndIP()
	if err != nil {
		hwlog.RunLog.Errorf("get current user or ip info failed: %s", err.Error())
		return util.ErrorExitCode
	}

	if installParam == nil {
		hwlog.RunLog.Error("installParam is nil")
		return util.ErrorExitCode
	}

	pathMgr, err := util.InitInstallDirPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("init install path mgr failed: %v", err)
		return util.ErrorExitCode
	}
	if err = util.CheckCurrentPath(pathMgr.GetWorkPath()); err != nil {
		fmt.Println("execute command failed")
		hwlog.RunLog.Error(err)
		return util.ErrorExitCode
	}
	if err = envutils.GetFlock(util.MefCenterLock).Lock(curController.getName()); err != nil {
		fmt.Println("execute command failed: the last is not complete")
		hwlog.RunLog.Error("execute command failed: the last is not complete, has not been unlocked yet")
		hwlog.OpLog.Errorf("[%s@%s] execute command failed", user, ip)
		return util.ErrorExitCode
	}
	defer envutils.GetFlock(util.MefCenterLock).Unlock()
	curController.setInstallParam(installParam)
	curController.printExecutingLog(ip, user)
	if err = curController.doControl(); err != nil {
		curController.printFailedLog(ip, user)
		if err.Error() == util.NotGenCertErrorStr {
			return util.NotGenCertErrorCode
		}
		return util.ErrorExitCode
	}
	curController.printSuccessLog(ip, user)
	return 0
}

func initLog(installParam *util.InstallParamJsonTemplate) error {
	if installParam == nil {
		return errors.New("installParam is nil")
	}
	logDirPath := installParam.LogDir
	logBackupDirPath := installParam.LogBackupDir
	logPathMgr := util.InitLogDirPathMgr(logDirPath, logBackupDirPath)
	logPath, err := fileutils.CheckOriginPath(logPathMgr.GetInstallLogPath())
	if err != nil {
		return fmt.Errorf("check log path %s failed:%s", logPath, err.Error())
	}
	logBackupPath, err := fileutils.CheckOriginPath(logPathMgr.GetInstallLogBackupPath())
	if err != nil {
		return fmt.Errorf("check log backup path %s failed:%s", logBackupPath, err.Error())
	}

	if err = util.InitLogPath(logPath, logBackupPath); err != nil {
		return fmt.Errorf("init log path %s failed:%s", logPath, err.Error())
	}
	return nil
}

func getOperateMap(operate string) map[string]controller {
	return map[string]controller{
		util.StartOperateFlag:     &operateController{operate: operate},
		util.StopOperateFlag:      &operateController{operate: operate},
		util.RestartOperateFlag:   &operateController{operate: operate},
		util.UninstallFlag:        &uninstallController{operate: operate},
		util.UpgradeFlag:          &upgradeController{operate: operate},
		util.ExchangeCaFlag:       &exchangeCertsController{operate: operate},
		util.UpdateKmcFlag:        &updateKmcController{operate: operate},
		util.ImportCrlFlag:        &importCrlController{operate: operate},
		util.ManageThirdComponent: &manageThirdComponent{operate: operate},
		util.AlarmCfgFlag:         &alarmCfgController{operate: operate},
		util.GetAlarmCfgFlag:      &getAlarmCfgController{operate: operate},
	}
}
