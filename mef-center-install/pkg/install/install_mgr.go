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

var (
	necessaryTools = [...]string{
		"sh",
		"kubectl",
		"docker",
		"uname",
		"cp",
		"grep",
		"useradd",
		"wc",
		"who",
	}
)

// SftInstallCtl is the main struct to install mef-center
type SftInstallCtl struct {
	optComponents            []string
	installPathMgr           *util.InstallDirPathMgr
	logPathMgr               *util.LogDirPathMgr
	compulsoryComponents     map[string]*util.InstallComponent
	optionalComponents       map[string]*util.InstallComponent
	fullInstallingComponents map[string]*util.InstallComponent
}

// DoInstall is the main function to install mef-center
func (sic *SftInstallCtl) DoInstall() error {
	var installTasks = []func() error{
		sic.init,
		sic.preCheck,
		sic.prepareMefUser,
		sic.prepareComponentLogDir,
		sic.prepareInstallPkgDir,
		sic.prepareCerts,
		sic.prepareWorkingDir,
		sic.prepareYaml,
		sic.setInstallJson,
		sic.componentsInstall,
	}

	defer sic.clearAll()
	for _, function := range installTasks {
		if err := function(); err != nil {
			return err
		}
	}

	return nil
}

func (sic *SftInstallCtl) init() error {
	sic.compulsoryComponents = util.GetCompulsoryMap()
	sic.optionalComponents = util.GetOptionalMap()
	if err := sic.setComponents(); err != nil {
		return nil
	}
	sic.setComponentVersion()
	sic.setFullComponents()

	return nil
}

func (sic *SftInstallCtl) preCheck() error {
	var checkTasks = []func() error{
		sic.checkUser,
		sic.checkInstalled,
		sic.checkArch,
		sic.checkDiskSpace,
		sic.checkNecessaryTools,
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
		hwlog.RunLog.Error("unsupported arch")
		return errors.New("unsupported arch")
	}

	hwlog.RunLog.Info("check Arch successful")
	return nil
}

func (sic *SftInstallCtl) checkDiskSpace() error {
	rootPath := sic.installPathMgr.GetRootPath()
	availSpace, err := util.GetDiskFree(rootPath)
	if err != nil {
		hwlog.RunLog.Errorf("get disk free space failed: %s", err.Error())
		return errors.New("get disk free space failed")
	}

	if availSpace < util.InstallDiskSpace {
		hwlog.RunLog.Error("no enough space to install mef-center")
		return errors.New("no enough space to install mef-center")
	}

	return nil
}

func (sic *SftInstallCtl) checkNecessaryTools() error {
	hwlog.RunLog.Info("start to check necessary tools")
	for _, tool := range necessaryTools {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("look path of [%s] failed, error: %s", tool, err.Error())
		}
	}
	hwlog.RunLog.Info("check necessary tools success")
	return nil
}

func (sic *SftInstallCtl) setComponents() error {
	for _, component := range sic.optComponents {
		if sic.optionalComponents[component] == nil {
			hwlog.RunLog.Errorf("unsupported component %s", component)
			return fmt.Errorf("unsupported component %s", component)
		}
		sic.optionalComponents[component].Required = true
	}

	return nil
}

func (sic *SftInstallCtl) setComponentVersion() {
	for _, component := range sic.optionalComponents {
		component.SetVersion()
	}
	for _, component := range sic.compulsoryComponents {
		component.SetVersion()
	}
}

func (sic *SftInstallCtl) setFullComponents() {
	sic.fullInstallingComponents = sic.compulsoryComponents
	for key, value := range sic.optionalComponents {
		if !value.Required {
			continue
		}
		sic.fullInstallingComponents[key] = value
	}
}

func (sic *SftInstallCtl) setInstallJson() error {
	hwlog.RunLog.Info("start to set install json")
	jsonHandler := util.InstallParamJsonTemplate{
		Components: sic.getFullComponents(),
		InstallDir: sic.installPathMgr.GetRootPath(),
		LogDir:     sic.logPathMgr.GetLogRootPath(),
	}
	jsonPath := sic.installPathMgr.WorkPathMgr.GetInstallParamJsonPath()
	if err := jsonHandler.SetInstallParamJsonInfo(jsonPath); err != nil {
		hwlog.RunLog.Errorf("record install_param.json failed: %v", err.Error())
		return err
	}
	hwlog.RunLog.Info("set install json successful")
	return nil
}

func (sic *SftInstallCtl) checkInstalled() error {
	hwlog.RunLog.Info("start to check if the software has been installed")
	_, err := os.Stat(sic.installPathMgr.GetMefPath())
	if err == nil {
		hwlog.RunLog.Error("the software has already been installed")
		return errors.New("the software has already been installed")
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

func (sic *SftInstallCtl) prepareComponentLogDir() error {
	hwlog.RunLog.Info("start to prepare components' log dir")
	for _, component := range sic.fullInstallingComponents {
		if err := (*component).PrepareLogDir(sic.logPathMgr); err != nil {
			return err
		}
	}
	hwlog.RunLog.Info("prepare components' log dir successful")
	return nil
}

func (sic *SftInstallCtl) prepareInstallPkgDir() error {
	hwlog.RunLog.Info("start to prepare install_package dir")
	installPkgDir := sic.installPathMgr.GetInstallPkgDir() + "/"
	if err := utils.MakeSureDir(installPkgDir); err != nil {
		hwlog.RunLog.Errorf("prepare install_package dir failed: %s", err.Error())
		return errors.New("prepare install_package dir failed")
	}
	hwlog.RunLog.Info("prepare install_package dir successful")
	return nil
}

func (sic *SftInstallCtl) prepareCerts() error {
	certHandleCtl := certPrepareCtl{
		certPathMgr: sic.installPathMgr.ConfigPathMgr,
		components:  sic.fullInstallingComponents,
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
	workingDirHandleCtl := workingDirCtl{
		pathMgr:     sic.installPathMgr.WorkPathAMgr,
		mefLinkPath: sic.installPathMgr.GetWorkPath(),
		components:  sic.fullInstallingComponents,
	}

	if err := workingDirHandleCtl.doPrepare(); err != nil {
		hwlog.RunLog.Errorf("prepare working dir failed: %v", err.Error())
		return err
	}
	return nil
}

func (sic *SftInstallCtl) prepareYaml() error {
	hwlog.RunLog.Info("start to prepare components' yaml")
	yamlDealers := GetYamlDealers(
		sic.fullInstallingComponents, sic.installPathMgr, sic.logPathMgr.GetLogRootPath())
	for _, yamlDealer := range yamlDealers {
		err := yamlDealer.EditSingleYaml(sic.getFullComponents())
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
	for _, component := range sic.fullInstallingComponents {
		if err := (*component).LoadAndSaveImage(sic.installPathMgr.WorkPathAMgr); err != nil {
			return fmt.Errorf("install component [%s] failed: %s", component.Name, err.Error())
		}

		if err := (*component).ClearDockerFile(sic.installPathMgr.WorkPathAMgr); err != nil {
			return fmt.Errorf("clear component [%s]'s docker file failed: %s",
				component.Name, err.Error())
		}

		if err := (*component).ClearLibDir(sic.installPathMgr.WorkPathAMgr); err != nil {
			return fmt.Errorf("clear component [%s]'s lib dir failed: %s",
				component.Name, err.Error())
		}
	}
	fmt.Println("prepare docker image success")
	hwlog.RunLog.Info("-----Install components successful-----")

	return nil
}

func (sic *SftInstallCtl) clearAll() {
	// todo 待实现
	return
}

func (sic *SftInstallCtl) getFullComponents() []string {
	var result []string
	for module := range sic.fullInstallingComponents {
		result = append(result, module)
	}
	return result
}

// GetSftInstallCtl is used to init a SftInstallCtl struct
func GetSftInstallCtl(optionalComponents []string, installPath string, logRootPath string) SftInstallCtl {
	return SftInstallCtl{
		optComponents:  optionalComponents,
		installPathMgr: util.InitInstallDirPathMgr(installPath),
		logPathMgr:     util.InitLogDirPathMgr(logRootPath),
	}
}
