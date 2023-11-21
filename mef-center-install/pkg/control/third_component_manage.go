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

	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// ManageThirdComponentFlow is used to
type ManageThirdComponentFlow struct {
	pathMgr   *util.InstallDirPathMgr
	component string
	operate   string
	SubParam
}

// SubParam sub parameter of manage third component
type SubParam struct {
	InstallPackagePath string
	InstallCmsPath     string
	InstallCrlPath     string
}

// NewThirdComponentManageFlow an ManageThirdComponentFlow struct
func NewThirdComponentManageFlow(component, operate string, subParams SubParam,
	pathMgr *util.InstallDirPathMgr) *ManageThirdComponentFlow {
	return &ManageThirdComponentFlow{
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
func (mtc *ManageThirdComponentFlow) DoManage() error {
	if err := mtc.checkParam(); err != nil {
		fmt.Println(err)
		return err
	}
	if mtc.component == util.IcsManagerName {
		ics := icsManager{pathMgr: mtc.pathMgr, name: util.IcsManagerName, operate: mtc.operate}
		return ics.operateIcsManager(mtc.SubParam)
	}

	return nil
}

func (mtc *ManageThirdComponentFlow) checkParam() error {
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
		if err := ics.install(args.InstallPackagePath, args.InstallCmsPath, args.InstallCrlPath); err != nil {
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

func (ics icsManager) install(tarPath, cmsPath, crlPath string) error {
	fmt.Printf("start to install %s\n", ics.name)
	icsExist := fileutils.IsExist(ics.pathMgr.GetIcsPath())
	exist, err := util.OptionComponentExist(util.IcsManagerName)
	if err != nil {
		return err
	}
	if exist && icsExist {
		fmt.Printf("%s has already been installed\n", ics.name)
		return fmt.Errorf("%s has already been installed", ics.name)
	}
	unpackTarPath, err := ics.prepareFile(tarPath, cmsPath, crlPath)
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
		hwlog.RunLog.Errorf("install %s error, get more information in ics-manager log", ics.name)
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

func (ics icsManager) prepareFile(tarPath, cmsPath, crlPath string) (string, error) {
	fmt.Println("start to verify ics-manager package")
	pathMap := map[string]string{
		"tar file": tarPath,
		"cms file": cmsPath,
		"crl file": crlPath,
	}
	for fileTag, filePath := range pathMap {
		if err := ics.checkInstallPaths(fileTag, filePath); err != nil {
			return "", err
		}
	}
	// when two input parameters are the same, the function can be used to check whether the CRL file is valid
	crlToUpdateValid, err := cmsverify.CompareCrls(crlPath, crlPath)
	if err != nil || int(crlToUpdateValid) != util.CompareSame {
		fmt.Println("crl file is invalid")
		hwlog.RunLog.Error("crl file is invalid")
		return "", errors.New("crl file is invalid")
	}
	if err = cmsverify.VerifyPackage(crlPath, cmsPath, tarPath); err != nil {
		fmt.Println("verify package failed")
		hwlog.RunLog.Errorf("verify package failed,error:%v", err)
		return "", errors.New("verify package failed")
	}

	hwlog.RunLog.Infof("verify %s package success", ics.name)
	unpackTarPath, err := ics.pathMgr.GetIcsTempTarPath()
	if err != nil {
		hwlog.RunLog.Errorf("unpack tar path check failed: %s", err.Error())
		return "", errors.New("unpack tar path check failed")
	}
	if err := fileutils.ExtraTarGzFile(tarPath, unpackTarPath, true); err != nil {
		hwlog.RunLog.Errorf("unzip tar file failed: %s", err.Error())
		return "", errors.New("unzip tar file failed")
	}
	return unpackTarPath, nil
}

func (ics icsManager) checkInstallPaths(fileTag, filePath string) error {
	const maxFileSize = 512

	if !fileutils.IsExist(filePath) {
		hwlog.RunLog.Errorf("%s does not exist", fileTag)
		return fmt.Errorf("%s does not exist", fileTag)
	}

	if _, err := fileutils.RealFileCheck(filePath, true, false, maxFileSize); err != nil {
		hwlog.RunLog.Errorf("check %s failed: %v", fileTag, err)
		return fmt.Errorf("check %s failed", fileTag)
	}

	return nil
}

func (ics icsManager) uninstall() error {
	if err := util.DeleteComponentToInstallInfo(util.IcsManagerName,
		ics.pathMgr.WorkPathMgr.GetInstallParamJsonPath()); err != nil {
		if strings.Contains(err.Error(), util.ComponentNotInstalled) {
			hwlog.RunLog.Errorf("%s not installed yet, cannot %s", ics.name, ics.operate)
			fmt.Printf("%s not installed yet, cannot %s\n", ics.name, ics.operate)
		}
		hwlog.RunLog.Errorf("delete ics from install param json failed, error: %v", err)
		return err
	}
	if exist := fileutils.IsExist(ics.pathMgr.GetIcsPath()); !exist {
		fmt.Printf("%s has uninstalled by others, start to clean residual files\n", ics.name)
		hwlog.RunLog.Infof("%s has uninstalled by others, start to clean residual files", ics.name)
		if err := fileutils.DeleteAllFileWithConfusion(ics.pathMgr.ConfigPathMgr.GetIcsCertDir()); err != nil {
			hwlog.RunLog.Errorf("when uninstall ics-manager, delete inner root failed: %v", err)
			return err
		}
		fmt.Println("clean residual files success")
		hwlog.RunLog.Info("clean residual files success")
		return nil
	}
	if err := fileutils.DeleteAllFileWithConfusion(ics.pathMgr.ConfigPathMgr.GetIcsCertDir()); err != nil {
		if err := util.AddComponentToInstallInfo(util.IcsManagerName,
			ics.pathMgr.WorkPathMgr.GetInstallParamJsonPath()); err != nil {
			hwlog.RunLog.Errorf("%s uninstall failed, rollback failed", ics.name)
		}
		hwlog.RunLog.Errorf("when uninstall ics-manager, delete inner root failed: %v", err)
		return err
	}
	if err := ics.Operate(); err != nil {
		if err := util.AddComponentToInstallInfo(util.IcsManagerName,
			ics.pathMgr.WorkPathMgr.GetInstallParamJsonPath()); err != nil {
			hwlog.RunLog.Errorf("%s uninstall failed, rollback failed", ics.name)
			return fmt.Errorf("%s uninstall failed, rollback failed", ics.name)
		}
	}
	return nil
}

// Operate use to operate ics manager in start, stop, restart
func (ics icsManager) Operate() error {
	if exist := fileutils.IsExist(ics.pathMgr.GetIcsPath()); !exist {
		fmt.Printf("%s has uninstalled by others, cannot %s\n", ics.name, ics.operate)
		hwlog.RunLog.Warnf("%s has uninstalled by others, cannot %s", ics.name, ics.operate)
		return nil
	}

	hwlog.RunLog.Infof("start to %s module %s", ics.operate, ics.name)
	fmt.Printf("start to %s module %s\n", ics.operate, ics.name)
	switch ics.operate {
	case util.StartOperateFlag:
		if err := ics.exchange(); err != nil {
			hwlog.RunLog.Errorf("%s %s when exchange ca failed: %v", ics.name, ics.operate, err)
			return err
		}
	case util.StopOperateFlag:
	case util.RestartOperateFlag:
		if err := ics.exchange(); err != nil {
			hwlog.RunLog.Errorf("%s %s when exchange ca failed: %v", ics.name, ics.operate, err)
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
	hwlog.RunLog.Infof("%s result:%s", ics.operate, res)
	fmt.Printf("%s module %s successful\n", ics.operate, ics.name)
	hwlog.RunLog.Infof("%s module %s successful", ics.operate, ics.name)
	return nil
}

func (ics icsManager) exchange() error {
	importPath := filepath.Join(ics.pathMgr.GetRootPath(), "/ICS-Manager/ics-config/root-ca/ics-cert/RootCA.crt")
	exportPath := filepath.Join(ics.pathMgr.GetRootPath(), "/ICS-Manager/ics-config/root-ca/mef-cert/RootCA.crt")
	// make sure exportPath file not exist
	err := fileutils.DeleteFile(exportPath)
	if err != nil {
		hwlog.RunLog.Errorf("failed to delete existing export cert file,%s", err.Error())
		return errors.New("failed to delete existing export cert file")
	}
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
