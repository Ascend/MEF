// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package install

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// SftInstallCtl is the main struct to install mef-center
type SftInstallCtl struct {
	util.SoftwareMgr
	logPathMgr          *util.LogDirPathMgr
	installedComponents []string
}

// DoInstall is the main function to install mef-center
func (sic *SftInstallCtl) DoInstall() error {
	// 校验失败时不进行环境清理
	if err := sic.preCheck(); err != nil {
		return err
	}

	var installTasks = []func() error{
		sic.prepareMefUser,
		sic.prepareComponentLogDir,
		sic.prepareComponentLogBackupDir,
		sic.prepareWorkingDir,
		sic.setInstallJson,
		sic.prepareK8sLabel,
		sic.prepareCerts,
		sic.prepareYaml,
		sic.componentsInstall,
		sic.setCenterMode,
	}

	for _, function := range installTasks {
		err := function()
		if err == nil {
			continue
		}

		sic.clearAll()
		return err
	}

	return nil
}

func (sic *SftInstallCtl) preCheck() error {
	var checkTasks = []func() error{
		sic.checkUser,
		sic.checkNecessaryTools,
		sic.checkArch,
		sic.checkInstalled,
		sic.checkDiskSpace,
	}

	for _, function := range checkTasks {
		if err := function(); err != nil {
			return err
		}
	}

	fmt.Println("install pre check success")
	return nil
}

func (sic *SftInstallCtl) checkUser() error {
	hwlog.RunLog.Info("start to check user")
	if err := util.CheckUser(); err != nil {
		fmt.Println("current user is not root, cannot install")
		hwlog.RunLog.Errorf("check user failed: %s", err.Error())
		return err
	}
	hwlog.RunLog.Info("check user successful")
	return nil
}

func (sic *SftInstallCtl) checkArch() error {
	hwlog.RunLog.Info("start to check Arch")
	arch, err := util.GetArch()
	if err != nil {
		hwlog.RunLog.Errorf("get arch info failed: %s", err.Error())
		return err
	}

	if arch != util.Arch64 && arch != util.X86 {
		fmt.Println("does not support the current CPU arch, cannot install")
		hwlog.RunLog.Error("unsupported arch")
		return errors.New("unsupported arch")
	}

	hwlog.RunLog.Info("check Arch successful")
	return nil
}

func (sic *SftInstallCtl) checkDiskSpace() error {
	if err := util.CheckDiskSpace(sic.InstallPathMgr.GetRootPath(), util.InstallDiskSpace); err != nil {
		hwlog.RunLog.Errorf("check install disk space failed: %s", err.Error())
		return errors.New("check install disk space failed")
	}

	return nil
}

func (sic *SftInstallCtl) checkNecessaryTools() error {
	hwlog.RunLog.Info("start to check necessary tools")
	for _, tool := range util.GetNecessaryTools() {
		if _, err := exec.LookPath(tool); err != nil {
			fmt.Printf("necessary tool %s does not exist, cannot install\n", tool)
			return fmt.Errorf("look path of [%s] failed, error: %s", tool, err.Error())
		}
	}

	if _, err := exec.LookPath(util.Haveged); err != nil {
		fmt.Printf("warning: [%s] not found, system may be slow to read random numbers without it\n", util.Haveged)
		hwlog.RunLog.Warnf("[%s] not found, system may be slow to read random numbers without it", util.Haveged)
	}
	hwlog.RunLog.Info("check necessary tools success")
	return nil
}

func (sic *SftInstallCtl) setInstallJson() error {
	hwlog.RunLog.Info("start to set install json")
	jsonHandler := util.InstallParamJsonTemplate{
		Components:   sic.Components,
		InstallDir:   sic.InstallPathMgr.GetRootPath(),
		LogDir:       sic.logPathMgr.GetLogRootPath(),
		LogBackupDir: sic.logPathMgr.GetLogBackupRootPath(),
	}
	jsonPath := sic.InstallPathMgr.WorkPathMgr.GetInstallParamJsonPath()
	if err := jsonHandler.SetInstallParamJsonInfo(jsonPath); err != nil {
		hwlog.RunLog.Errorf("record install_param.json failed: %v", err.Error())
		return err
	}
	hwlog.RunLog.Info("set install json successful")
	return nil
}

func (sic *SftInstallCtl) checkInstalled() error {
	hwlog.RunLog.Info("start to check if the software has been installed")
	_, err := os.Stat(sic.InstallPathMgr.GetMefPath())
	if err == nil {
		hwlog.RunLog.Error("the software has already been installed")
		fmt.Println("MEF-Center has already been installed")
		return errors.New("the software has already been installed")
	}

	k8sMgr := util.K8sLabelMgr{}
	exists, err := k8sMgr.CheckK8sLabel()
	if err != nil {
		return err
	}
	if exists {
		hwlog.RunLog.Error("the software has already been installed since k8s label exists")
		fmt.Println("mef-center-node label exists, that MEF-Center might have already been installed")
		return errors.New("the software has already been installed since k8s label exists")
	}
	hwlog.RunLog.Info("check if the software has been installed successful")
	return nil
}

func (sic *SftInstallCtl) prepareMefUser() error {
	hwlog.RunLog.Info("start to prepare mef user")

	usrMgr := common.NewUserMgr(util.MefCenterName, util.MefCenterGroup, util.MefCenterUid, util.MefCenterGid)
	if err := usrMgr.AddUserAccount(); err != nil {
		hwlog.RunLog.Errorf("prepare mef user failed:%s", err.Error())
		return err
	}

	hwlog.RunLog.Info("prepare mef user successful")
	return nil
}

func (sic *SftInstallCtl) prepareK8sLabel() error {
	hwlog.RunLog.Info("start to set label for master node")
	k8sMgr := util.K8sLabelMgr{}
	if err := k8sMgr.PrepareK8sLabel(); err != nil {
		return err
	}
	hwlog.RunLog.Info("start to set label for master node")
	return nil
}

func (sic *SftInstallCtl) prepareComponentLogDir() error {
	hwlog.RunLog.Info("start to prepare components' log dir")
	for _, component := range sic.Components {
		componentMgr := util.GetComponentMgr(component)
		if err := componentMgr.PrepareLogDir(sic.logPathMgr); err != nil {
			return err
		}
	}
	hwlog.RunLog.Info("prepare components' log dir successful")
	return nil
}

func (sic *SftInstallCtl) prepareComponentLogBackupDir() error {
	hwlog.RunLog.Info("start to prepare components' log backup dir")
	for _, component := range sic.Components {
		componentMgr := util.GetComponentMgr(component)
		if err := componentMgr.PrepareLogBackupDir(sic.logPathMgr); err != nil {
			return err
		}
	}
	hwlog.RunLog.Info("prepare components' log backup dir successful")
	return nil
}

func (sic *SftInstallCtl) prepareInstallPkgDir() error {
	hwlog.RunLog.Info("start to prepare install_package dir")
	installPkgDir := sic.InstallPathMgr.GetInstallPkgDir() + "/"
	if err := utils.MakeSureDir(installPkgDir); err != nil {
		hwlog.RunLog.Errorf("prepare install_package dir failed: %s", err.Error())
		return errors.New("prepare install_package dir failed")
	}
	hwlog.RunLog.Info("prepare install_package dir successful")
	return nil
}

func (sic *SftInstallCtl) prepareCerts() error {
	certHandleCtl := certPrepareCtl{
		certPathMgr: sic.InstallPathMgr.ConfigPathMgr,
		components:  sic.Components,
	}

	hwlog.RunLog.Info("-----Start to prepare certs-----")
	if err := certHandleCtl.doPrepare(); err != nil {
		hwlog.RunLog.Errorf("prepare certs failed: %v", err.Error())
		return err
	}
	hwlog.RunLog.Info("-----Prepare certs successful-----")
	return nil
}

func (sic *SftInstallCtl) prepareWorkingDir() error {
	workingDirHandleCtl := GetWorkingDirMgr(
		sic.InstallPathMgr.WorkPathAMgr,
		sic.InstallPathMgr.GetWorkPath(),
		sic.Components)

	if err := workingDirHandleCtl.DoInstallPrepare(); err != nil {
		hwlog.RunLog.Errorf("prepare working dir failed: %v", err.Error())
		return err
	}
	return nil
}

func (sic *SftInstallCtl) prepareYaml() error {
	hwlog.RunLog.Info("start to prepare components' yaml")
	for _, component := range sic.Components {
		yamlPath := sic.InstallPathMgr.WorkPathAMgr.GetComponentYamlPath(component)
		yamlDealer := GetYamlDealer(sic.InstallPathMgr, component, sic.logPathMgr.GetLogRootPath(),
			sic.logPathMgr.GetLogBackupRootPath(), yamlPath)

		err := yamlDealer.EditSingleYaml(sic.Components)
		if err != nil {
			return err
		}
	}
	hwlog.RunLog.Info("prepare components' yaml successful")
	return nil
}

func (sic *SftInstallCtl) componentsInstall() error {
	fmt.Println("start to prepare docker image")
	hwlog.RunLog.Info("-----Start to install components-----")
	for _, component := range sic.Components {
		componentMgr := util.GetComponentMgr(component)
		if err := componentMgr.LoadAndSaveImage(sic.InstallPathMgr.WorkPathAMgr); err != nil {
			return fmt.Errorf("install component [%s] failed: %s", component, err.Error())
		}
		sic.installedComponents = append(sic.installedComponents, component)

		if err := componentMgr.ClearDockerFile(sic.InstallPathMgr.WorkPathAMgr); err != nil {
			return fmt.Errorf("clear component [%s]'s docker file failed: %s", component, err.Error())
		}

		if err := componentMgr.ClearLibDir(sic.InstallPathMgr.WorkPathAMgr); err != nil {
			return fmt.Errorf("clear component [%s]'s lib dir failed: %s", component, err.Error())
		}
	}
	fmt.Println("prepare docker image success")
	hwlog.RunLog.Info("-----Install components successful-----")
	return nil
}

func (sic *SftInstallCtl) setCenterMode() error {
	hwlog.RunLog.Info("-----Start to set mef-center mode-----")
	modeMgr := util.GetCenterModeMgr(sic.InstallPathMgr)
	if err := modeMgr.SetWorkDirMode(); err != nil {
		fmt.Println("set work dir mode failed")
		hwlog.RunLog.Errorf("set work dir mode failed: %s", err.Error())
		return errors.New("set work dir mode failed")
	}

	if err := modeMgr.SetConfigDirMode(); err != nil {
		fmt.Println("set config dir mode failed")
		hwlog.RunLog.Errorf("set config dir mode failed: %s", err.Error())
		return errors.New("set config dir mode failed")
	}
	hwlog.RunLog.Info("-----set mef-center mode success-----")
	return nil
}

func (sic *SftInstallCtl) clearAll() {
	fmt.Println("install failed, start to clear environment")
	hwlog.RunLog.Info("-----Start to clear environment-----")
	if err := sic.ClearDockerImage(sic.installedComponents); err != nil {
		fmt.Println("clear environment failed, please clear manually")
		hwlog.RunLog.Warnf("clear environment meets err:%s, need to do it manually", err.Error())
		hwlog.RunLog.Info("-----End to clear environment-----")
		return
	}

	if err := sic.ClearAndLabel(); err != nil {
		fmt.Println("clear environment failed, please clear manually")
		hwlog.RunLog.Warnf("clear environment meets err:%s, need to do it manually", err.Error())
		hwlog.RunLog.Info("-----End to clear environment-----")
		return
	}
	fmt.Println("clear environment success")
	hwlog.RunLog.Info("-----End to clear environment-----")
	return
}

// GetSftInstallMgrIns is used to init a SftInstallCtl struct
func GetSftInstallMgrIns(components []string,
	installPath string, logRootPath string, logBackupRootPath string) *SftInstallCtl {
	return &SftInstallCtl{
		SoftwareMgr: util.SoftwareMgr{
			Components:     components,
			InstallPathMgr: util.InitInstallDirPathMgr(installPath),
		},
		logPathMgr: util.InitLogDirPathMgr(logRootPath, logBackupRootPath),
	}
}
