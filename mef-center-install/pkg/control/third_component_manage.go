// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package control

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/mef/common/cmsverify"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// ManageThridComponentFlow is used to
type ManageThridComponentFlow struct {
	pathMgr   *util.InstallDirPathMgr
	component string
	operate   string
	SubParam
}

// SubParam sub parameter of manage third component
type SubParam struct {
	InstallPackagePath string
}

// NewThirdComponentManageFlow an ManageThridComponentFlow struct
func NewThirdComponentManageFlow(component, operate string, subParams SubParam,
	pathMgr *util.InstallDirPathMgr) *ManageThridComponentFlow {
	return &ManageThridComponentFlow{
		pathMgr:   pathMgr,
		component: component,
		operate:   operate,
		SubParam:  subParams,
	}
}

func thirdComponent() []string {
	return []string{
		util.IcsManagerName,
	}
}

func operateThirdComponent() []string {
	return []string{
		util.OperateInstall,
		util.OperateUninstall,
	}
}

// DoManage is the main func to manage third component
func (ecf *ManageThridComponentFlow) DoManage() error {
	if err := ecf.checkParam(); err != nil {
		return err
	}
	if ecf.component == util.IcsManagerName {
		ics := icsManager{pathMgr: ecf.pathMgr, name: util.IcsManagerName, operate: ecf.operate}
		return ics.operateIcsManager(ecf.SubParam)
	}

	return nil
}

func (mtc *ManageThridComponentFlow) checkParam() error {
	if err := util.CheckParamOption(thirdComponent(), mtc.component); err != nil {
		hwlog.RunLog.Errorf("check parameter component error, %v", err)
		return fmt.Errorf("check parameter component error, %v", err)
	}
	if err := util.CheckParamOption(operateThirdComponent(), mtc.operate); err != nil {
		hwlog.RunLog.Errorf("check parameter operate error, %v", err)
		return fmt.Errorf("check parameter operate error, %v", err)
	}

	return nil
}

type icsManager struct {
	name    string
	operate string
	pathMgr *util.InstallDirPathMgr
}

func (ics icsManager) operateIcsManager(args SubParam) error {
	switch ics.operate {
	case util.OperateInstall:
		if err := ics.install(args.InstallPackagePath); err != nil {
			util.ClearPakEnv(ics.pathMgr.WorkPathMgr.GetVarDirPath())
			return err
		}
		return nil
	case util.OperateUninstall:
		return ics.uninstall()
	default:
		return errors.New("not support operate")
	}
}

func (ics icsManager) install(zipPath string) error {
	fmt.Printf("start to install %s\n", ics.name)
	exist, err := util.OptionComponentExist(util.IcsManagerName)
	if err != nil {
		return err
	}
	if exist {
		fmt.Printf("%s has already been installed\n", ics.name)
		return fmt.Errorf("%s has already been installed", ics.name)
	}
	unpackTarPath, err := ics.prepareFile(zipPath)
	if err != nil {
		return err
	}

	installPath, err := fileutils.CheckOwnerAndPermission(
		filepath.Join(unpackTarPath, util.InstallDirName, "install.sh"), util.ModeUmask0277, 0)
	if err != nil {
		hwlog.RunLog.Errorf("check %s install.sh owner and permission failed: %v", ics.name, err)
		return err
	}

	installInfo, err := util.GetInstallInfo()
	if err != nil {
		return fmt.Errorf("get install info from json error:%v", err)
	}
	if err := util.AddComponentToInstallInfo(util.IcsManagerName,
		ics.pathMgr.WorkPathMgr.GetInstallParamJsonPath()); err != nil {
		return fmt.Errorf("add %s to installInfo json error: %v", ics.name, err)
	}

	res, err := envutils.RunCommand(installPath, envutils.DefCmdTimeoutSec, "-install_mode=dependent",
		fmt.Sprintf("-install_path=%s", installInfo.InstallDir),
		fmt.Sprintf("-log_path=%s", installInfo.LogDir),
		fmt.Sprintf("-log_backup_path=%s", installInfo.LogBackupDir))

	if err != nil {
		fmt.Printf("install %s failed\n", ics.name)
		hwlog.RunLog.Errorf("install %s error: %v, get more information in ics-manager log", ics.name, err)
		if err := util.DeleteComponentToInstallInfo(util.IcsManagerName,
			ics.pathMgr.WorkPathMgr.GetInstallParamJsonPath()); err != nil {
			return fmt.Errorf("install failed, delete %s from installInfo json error: %v", ics.name, err)
		}
		return errors.New("install failed")
	}
	fmt.Printf("install %s successful\n", ics.name)
	hwlog.RunLog.Infof("install %s, result %s", ics.name, res)
	return nil
}

func (ics icsManager) prepareFile(zipPath string) (string, error) {
	fmt.Println("start to verify ics-manager package")
	unpackPath := ics.pathMgr.GetRealVarDirPath()
	if err := util.CheckZipFile(unpackPath, zipPath); err != nil {
		hwlog.RunLog.Errorf("check %s zip path error: %v", ics.name, err)
		return "", fmt.Errorf("check %s zip path error: %v", ics.name, err)
	}
	unpackZipPath := ics.pathMgr.GetIcsTempZipPath()
	if err := common.ExtraUpgradeZipFile(zipPath, unpackZipPath); err != nil {
		hwlog.RunLog.Errorf("unzip zip file failed: %s", err.Error())
		return "", errors.New("unzip zip file failed")
	}
	zipContent, err := util.GetVerifyFileName(unpackZipPath, common.IcsManagerFlag)
	if err != nil {
		return "", err
	}
	// when two input parameters are the same, the function can be used to check whether the CRL file is valid
	crlToUpdateValid, err := cmsverify.CompareCrls(zipContent.CrlPath, zipContent.CrlPath)
	if err != nil || int(crlToUpdateValid) != util.CompareSame {
		fmt.Println("crl file is invalid")
		hwlog.RunLog.Error("crl file is invalid")
		return "", errors.New("crl file is invalid")
	}
	if err = cmsverify.VerifyPackage(zipContent.CrlPath, zipContent.CmsPath, zipContent.TarPath); err != nil {
		fmt.Println("verify package failed, the zip file might be tampered")
		hwlog.RunLog.Errorf("verify package failed,error:%v", err)
		return "", errors.New("verify package failed")
	}

	hwlog.RunLog.Infof("verify %s package success", ics.name)
	unpackTarPath := ics.pathMgr.GetIcsTempTarPath()
	if err := common.ExtraTarGzFile(zipContent.TarPath, unpackTarPath, true); err != nil {
		hwlog.RunLog.Errorf("unzip tar file failed: %s", err.Error())
		return "", errors.New("unzip tar file failed")
	}
	return unpackTarPath, nil
}

func (ics icsManager) uninstall() error {
	if err := util.DeleteComponentToInstallInfo(util.IcsManagerName,
		ics.pathMgr.WorkPathMgr.GetInstallParamJsonPath()); err != nil {
		if strings.Contains(err.Error(), util.ComponentNotInstalled) {
			fmt.Printf("%s not installed yet, cannot %s\n", ics.name, ics.operate)
		}
		return err
	}
	if err := fileutils.DeleteAllFileWithConfusion(ics.pathMgr.ConfigPathMgr.GetIcsCertDir()); err != nil {
		hwlog.RunLog.Errorf("when uninstall ics-manager, delete inner root failed: %v", err)
		return err
	}
	return ics.Operate()
}

// Operate use to operate ics manager in start, stop, restart
func (ics icsManager) Operate() error {
	hwlog.RunLog.Infof("start to %s module %s", ics.operate, ics.name)
	fmt.Printf("start to %s module %s\n", ics.operate, ics.name)

	switch ics.operate {
	case util.StartOperateFlag:
		if err := ics.exchange(); err != nil {
			hwlog.RunLog.Errorf("%s %s when exchage ca failed:%v", ics.name, ics.operate, err)
			return err
		}
	case util.StopOperateFlag:
	case util.RestartOperateFlag:
		if err := ics.exchange(); err != nil {
			hwlog.RunLog.Errorf("%s %s when exchage ca failed:%v", ics.name, ics.operate, err)
			return err
		}
	case util.UninstallFlag:
	default:
		hwlog.RunLog.Errorf("unsupported Operate type")
		return errors.New("unsupported Operate type")
	}

	runPath, err := fileutils.CheckOwnerAndPermission(ics.pathMgr.GetIcsRunPath(), util.ModeUmask0277, 0)
	if err != nil {
		hwlog.RunLog.Errorf("check %s run path failed: %v", ics.name, err)
		return err
	}
	res, err := envutils.RunCommand(runPath, envutils.DefCmdTimeoutSec, ics.operate)
	if err != nil {
		hwlog.RunLog.Errorf("%s component %s failed: %v", ics.operate, ics.name, err)
		return err
	}
	hwlog.RunLog.Infof("%s result:%v", ics.operate, res)
	fmt.Printf("%s module %s successful\n", ics.operate, ics.name)
	hwlog.RunLog.Infof("%s module %s successful", ics.operate, ics.name)
	return nil
}

func (ics icsManager) exchange() error {
	importPath := filepath.Join(ics.pathMgr.GetRootPath(), "/ICS-Manager/ics-config/root-ca/ics-cert/RootCA.crt")
	exportPath := filepath.Join(ics.pathMgr.GetRootPath(), "/ICS-Manager/ics-config/root-ca/mef-cert/RootCA.crt")
	exchangeFlow, err := NewExchangeCaFlow(importPath, exportPath, util.IcsManagerName, ics.pathMgr)
	if err != nil {
		return err
	}
	if err = exchangeFlow.DoExchange(); err != nil {
		hwlog.RunLog.Errorf("execute exchange flow failed: %s", err.Error())
		return err
	}
	return nil
}
