// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package control

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/install"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// UpgradePostFlowMgr is a struct that used to uninstall mef-center
type UpgradePostFlowMgr struct {
	util.SoftwareMgr
	logPathMgr             *util.LogDirPathMgr
	startedComponents      []string
	notInstalledComponents []string
	step                   int
}

// GetUpgradePostMgr is a func to init an UpgradePostFlowMgr struct
func GetUpgradePostMgr(components []string, installInfo *util.InstallParamJsonTemplate) (*UpgradePostFlowMgr, error) {
	pathMgr, err := util.InitInstallDirPathMgr(installInfo.InstallDir)
	if err != nil {
		return nil, fmt.Errorf("init upgrade post mgr failed: %v", err)
	}
	return &UpgradePostFlowMgr{
		SoftwareMgr: util.SoftwareMgr{
			Components:     components,
			InstallPathMgr: pathMgr,
		},
		logPathMgr: util.InitLogDirPathMgr(installInfo.LogDir, installInfo.LogBackupDir),
		step:       util.ClearUnpackPathStep,
	}, nil
}

// DoUpgrade is the main flow-control func to exec upgrade flow on the new package
func (upf *UpgradePostFlowMgr) DoUpgrade() error {
	var installTasks = []func() error{
		upf.checkVersion,
		upf.checkNecessaryTools,
		util.CheckDependentImage,
		upf.prepareK8sLabel,
		upf.createFlag,
		upf.prepareWorkCDir,
		upf.prepareLogDumpDir,
		upf.prepareCerts,
		upf.prepareYaml,
		upf.smoothUpgrade,
		upf.recordStarted,
		upf.deleteNameSpace,
		upf.removeDockerImage,
		upf.buildNewImage,
		upf.startNewPod,
		upf.resetSoftLink,
		upf.setCenterMode,
		upf.clearFlag,
	}

	for _, function := range installTasks {
		err := function()
		if err == nil {
			continue
		}

		upf.clearEnv()
		return err
	}

	return nil
}

func (upf *UpgradePostFlowMgr) checkVersion() error {
	hwlog.RunLog.Info("start to compare version")
	currentPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		hwlog.RunLog.Errorf("get current path failed: %s", err.Error())
		return errors.New("get current path failed")
	}

	newXmlPath := path.Join(filepath.Dir(currentPath), util.VersionXml)
	newVersion, err := GetVersion(newXmlPath)
	if err != nil {
		hwlog.RunLog.Errorf("get new version failed: %s", err.Error())
		return errors.New("get new version failed")
	}

	oldXmlPath, err := filepath.EvalSymlinks(upf.InstallPathMgr.WorkPathMgr.GetVersionXmlPath())
	if err != nil {
		hwlog.RunLog.Errorf("get old version.xml's abs path failed: %s", err.Error())
		return errors.New("get old version.xml's abs path failed")
	}

	oldVersion, err := GetVersion(oldXmlPath)
	if err != nil {
		hwlog.RunLog.Errorf("get old version failed: %s", err.Error())
		return errors.New("get old version failed")
	}

	ret, err := upf.compareVersion(newVersion, oldVersion)
	if err != nil {
		hwlog.RunLog.Errorf("compare versions failed: %s", err.Error())
		return errors.New("compare versions failed")
	}
	if !ret {
		hwlog.RunLog.Error("cannot upgrade to an older version")
		return errors.New("cannot upgrade to an older version")
	}

	hwlog.RunLog.Info("compare version succeeds")
	return nil
}

func (upf *UpgradePostFlowMgr) compareVersion(newVersion, oldVersion string) (bool, error) {
	oldNums := strings.Split(oldVersion, ".")
	newNums := strings.Split(newVersion, ".")

	for i := 0; i < len(oldNums) && i < len(newNums); i++ {
		oldNum, err := strconv.Atoi(oldNums[i])
		if err != nil {
			return false, err
		}
		newNum, err := strconv.Atoi(newNums[i])
		if err != nil {
			return false, err
		}
		if oldNum == newNum {
			continue
		}
		return newNum > oldNum, nil
	}
	return len(newNums) >= len(oldNums), nil
}

func (upf *UpgradePostFlowMgr) checkNecessaryTools() error {
	hwlog.RunLog.Info("start to check necessary tools")
	for _, tool := range util.GetNecessaryTools() {
		if _, err := exec.LookPath(tool); err != nil {
			hwlog.RunLog.Errorf("look path of [%s] failed, error: %s", tool, err.Error())
			return errors.New("necessary tools does not exists")
		}
	}

	if _, err := exec.LookPath(util.Haveged); err != nil {
		fmt.Printf("warning: [%s] not found, system may be slow to read random numbers without it\n", util.Haveged)
		hwlog.RunLog.Warnf("[%s] not found, system may be slow to read random numbers without it", util.Haveged)
	}
	hwlog.RunLog.Info("check necessary tools succeeds")
	return nil
}

func (upf *UpgradePostFlowMgr) prepareK8sLabel() error {
	hwlog.RunLog.Info("start to prepare label for master node")
	k8sMgr := util.K8sLabelMgr{}
	exists, err := k8sMgr.CheckK8sLabel()
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	err = k8sMgr.PrepareK8sLabel()
	if err != nil {
		return err
	}
	hwlog.RunLog.Info("set label for master node succeeds")
	return nil
}

func (upf *UpgradePostFlowMgr) createFlag() error {
	if err := fileutils.CreateFile(upf.InstallPathMgr.WorkPathMgr.GetUpgradeFlagPath(), fileutils.Mode400); err != nil {
		hwlog.RunLog.Errorf("create upgrade-flag failed: %s", err.Error())
		return errors.New("create upgrade-flag failed")
	}
	return nil
}

func (upf *UpgradePostFlowMgr) prepareWorkCDir() error {
	upf.step = util.ClearTempUpgradePathStep
	if err := fileutils.CreateDir(upf.InstallPathMgr.TmpPathMgr.GetWorkPath(), fileutils.Mode700); err != nil {
		hwlog.RunLog.Errorf("make sure temp upgrade dir failed: %s", err.Error())
		return errors.New("make sure temp upgrade dir failed")
	}

	workingDirHandleCtl := install.GetWorkingDirMgr(upf.InstallPathMgr.TmpPathMgr,
		upf.InstallPathMgr.GetWorkPath(), upf.Components)

	if err := workingDirHandleCtl.DoUpgradePrepare(); err != nil {
		hwlog.RunLog.Errorf("prepare working dir failed: %v", err.Error())
		return errors.New("prepare working dir failed")
	}

	if err := fileutils.DeleteAllFileWithConfusion(upf.InstallPathMgr.WorkPathMgr.GetVarDirPath()); err != nil {
		hwlog.RunLog.Errorf("delete unpack dir failed: %s", err.Error())
		return errors.New("delete unpack dir failed")
	}
	return nil
}

func (upf *UpgradePostFlowMgr) prepareLogDumpDir() error {
	hwlog.RunLog.Info("start to prepare log dump dir")
	if err := util.PrepareLogDumpDir(); err != nil {
		hwlog.RunLog.Errorf("prepare log dump dir failed, %v", err)
		return fmt.Errorf("prepare log dump dir failed, %v", err)
	}
	hwlog.RunLog.Info("prepare log dump dir successful")
	return nil
}

func (upf *UpgradePostFlowMgr) prepareYaml() error {
	hwlog.RunLog.Info("start to prepare components' yaml")
	for _, component := range upf.Components {
		yamlPath := upf.InstallPathMgr.TmpPathMgr.GetComponentYamlPath(component)
		yamlDealer := install.GetYamlDealer(upf.InstallPathMgr, component, upf.logPathMgr.GetLogRootPath(),
			upf.logPathMgr.GetLogBackupRootPath(), yamlPath)

		err := yamlDealer.EditSingleYaml(upf.Components)
		if err != nil {
			return err
		}
	}
	yamlPath := upf.InstallPathMgr.TmpPathMgr.GetComponentYamlPath(util.EdgeManagerName)
	if err := util.ModifyEndpointYaml(util.GetApiserverEndpoint(), yamlPath); err != nil {
		return err
	}
	hwlog.RunLog.Info("prepare components' yaml successful")
	return nil
}

func (upf *UpgradePostFlowMgr) smoothUpgrade() error {
	hwlog.RunLog.Info("start to smooth")
	upgradeSmoother, err := GetSmoother(upgradeFlow, upf.InstallPathMgr, upf.logPathMgr)
	if err != nil {
		return fmt.Errorf("get smoother failed: %s", err.Error())
	}
	if err = upgradeSmoother.smooth(); err != nil {
		return err
	}
	hwlog.RunLog.Info("smooth success")
	return nil
}

func (upf *UpgradePostFlowMgr) prepareCerts() error {
	hwlog.RunLog.Info("-----start to prepare certs-----")
	originalDir := upf.InstallPathMgr.ConfigPathMgr.GetKubeConfigCertDirPath()
	backupDir := originalDir + "_bak"
	if fileutils.IsLexist(originalDir) {
		if err := fileutils.RenameFile(originalDir, backupDir); err != nil {
			hwlog.RunLog.Errorf("rename kube config cert dir failed, %v", err)
			return err
		}
	}
	var successFlag bool
	defer func() {
		dirtToClean := originalDir
		if successFlag {
			dirtToClean = backupDir
		}
		if err := fileutils.DeleteAllFileWithConfusion(dirtToClean); err != nil {
			hwlog.RunLog.Errorf("clean kube config cert dir failed, %v", err)
		}
		if successFlag {
			return
		}
		if fileutils.IsLexist(backupDir) {
			if err := fileutils.RenameFile(backupDir, originalDir); err != nil {
				hwlog.RunLog.Errorf("restore kube config cert dir failed, %v", err)
			}
		}
	}()

	if err := util.PrepareKubeConfigCert(upf.InstallPathMgr.ConfigPathMgr); err != nil {
		return err
	}
	mefUid, mefGid, err := util.GetMefId()
	if err != nil {
		hwlog.RunLog.Errorf("get mef uid or gid failed: %s", err.Error())
		return errors.New("get mef uid or gid failed")
	}

	param := fileutils.SetOwnerParam{
		Path:       originalDir,
		Uid:        mefUid,
		Gid:        mefGid,
		Recursive:  true,
		IgnoreFile: false,
	}
	if err = fileutils.SetPathOwnerGroup(param); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner and group failed: %v", originalDir, err.Error())
		return errors.New("set kube config cert root path owner and group failed")
	}

	successFlag = true
	hwlog.RunLog.Info("-----prepare certs successful-----")
	return nil
}

func (upf *UpgradePostFlowMgr) recordStarted() error {
	hwlog.RunLog.Info("start to record started components")
	for _, c := range upf.SoftwareMgr.Components {
		dealer := &util.CtlComponent{
			Name:           c,
			InstallPathMgr: upf.SoftwareMgr.InstallPathMgr.WorkPathMgr,
		}
		started, err := util.CheckStarted(dealer.Name)
		if err != nil {
			hwlog.RunLog.Errorf("check component %s's status failed: %s", c, err.Error())
			return fmt.Errorf("check component %s's status failed", c)
		}

		if !started {
			continue
		}
		hwlog.RunLog.Infof("component %s is running", c)
		upf.startedComponents = append(upf.startedComponents, c)
	}

	for _, c := range util.GetAddedComponent() {
		dockerDealer := util.GetAscendDockerDealer(c, util.DockerTag)
		ret, err := dockerDealer.CheckImageExists()
		if err != nil {
			hwlog.RunLog.Errorf("check component %s's image failed: %s", c, err.Error())
			return fmt.Errorf("check component %s's image failed", c)
		}

		if ret {
			continue
		}

		upf.notInstalledComponents = append(upf.notInstalledComponents, c)
	}

	if len(upf.notInstalledComponents)+len(upf.startedComponents) == len(upf.Components) {
		upf.startedComponents = append(upf.startedComponents, upf.notInstalledComponents...)
	}
	hwlog.RunLog.Info("record started components succeeds")
	return nil
}

func (upf *UpgradePostFlowMgr) deleteNameSpace() error {
	upf.step = util.RestartPodStep
	return upf.ClearMEFCenterNamespace()
}

func (upf *UpgradePostFlowMgr) removeDockerImage() error {
	upf.step = util.LoadOldDockerStep
	for _, component := range upf.SoftwareMgr.Components {
		dockerDealerIns := util.GetAscendDockerDealer(component, util.DockerTag)
		if err := dockerDealerIns.DeleteImage(); err != nil {
			return err
		}
	}

	return nil
}

func (upf *UpgradePostFlowMgr) buildNewImage() error {
	hwlog.RunLog.Info("start to build new docker image")
	fmt.Println("start to build new docker image")
	upf.step = util.RemoveDockerStep
	for _, component := range upf.SoftwareMgr.Components {
		componentMgr := util.GetComponentMgr(component)
		if err := componentMgr.LoadAndSaveImage(upf.InstallPathMgr.TmpPathMgr); err != nil {
			hwlog.RunLog.Errorf("build component [%s]'s docker image failed: %s", component, err.Error())
			return fmt.Errorf("build component [%s]'s docker image failed", component)
		}

		if err := componentMgr.ClearDockerFile(upf.InstallPathMgr.TmpPathMgr); err != nil {
			hwlog.RunLog.Errorf("clear component [%s]'s docker file failed: %s", component, err.Error())
			return fmt.Errorf("clear component [%s]'s docker file failed", component)
		}

		if err := componentMgr.ClearLibDir(upf.InstallPathMgr.TmpPathMgr); err != nil {
			hwlog.RunLog.Errorf("clear component [%s]'s lib dir failed: %s", component, err.Error())
			return fmt.Errorf("clear component [%s]'s lib dir failed", component)
		}
	}
	hwlog.RunLog.Info("build new docker image succeeds")
	fmt.Println("build new docker image succeeds")
	return nil
}

func (upf *UpgradePostFlowMgr) startNewPod() error {
	hwlog.RunLog.Info("start to start pods")
	fmt.Println("start to start pods")
	upf.step = util.ClearNameSpaceStep
	for _, c := range upf.startedComponents {
		dealer := &util.CtlComponent{
			Name:           c,
			Operation:      util.StartOperateFlag,
			InstallPathMgr: upf.SoftwareMgr.InstallPathMgr.TmpPathMgr,
		}
		err := dealer.Operate()
		if err != nil {
			hwlog.RunLog.Errorf("start component %s failed", c)
			return fmt.Errorf("start component %s failed", c)
		}
	}

	hwlog.RunLog.Info("start pods succeeds")
	fmt.Println("start pods succeeds")
	return nil
}

func (upf *UpgradePostFlowMgr) resetSoftLink() error {
	backPath, err := upf.InstallPathMgr.GetTargetWorkPath()
	if err != nil {
		hwlog.RunLog.Errorf("get target work path failed: %s", err.Error())
		return err
	}

	if err = fileutils.DeleteAllFileWithConfusion(backPath); err != nil {
		hwlog.RunLog.Errorf("delete backUp path failed: %s", err.Error())
		return err
	}

	if err = fileutils.RenameFile(upf.InstallPathMgr.GetTmpUpgradePath(), backPath); err != nil {
		hwlog.RunLog.Errorf("rename temp-upgrade dir failed: %s", err.Error())
		return err
	}

	if fileutils.IsExist(upf.InstallPathMgr.GetWorkPath()) {
		if err = fileutils.DeleteFile(upf.InstallPathMgr.GetWorkPath()); err != nil {
			hwlog.RunLog.Errorf("remove old software dir symlink failed, error: %s", err.Error())
			return errors.New("remove old software dir symlink failed")
		}
	}

	if err = common.CreateSoftLink(backPath, upf.InstallPathMgr.GetWorkPath()); err != nil {
		hwlog.RunLog.Errorf("create software dir symlink failed, error: %s", err.Error())
		return errors.New("create software dir symlink failed")
	}

	if err = upf.InstallPathMgr.Reset(); err != nil {
		hwlog.RunLog.Errorf("reset path mgr failed: %v", err)
		return errors.New("reset path mgr failed")
	}

	return nil
}

func (upf *UpgradePostFlowMgr) setCenterMode() error {
	hwlog.RunLog.Info("-----Start to set mef-center mode-----")
	modeMgr := util.GetCenterModeMgr(upf.InstallPathMgr)
	if err := modeMgr.SetWorkDirMode(); err != nil {
		fmt.Println("set config dir mode failed")
		hwlog.RunLog.Errorf("set config dir mode failed: %s", err.Error())
		return errors.New("set config dir mode failed")
	}

	if err := modeMgr.SetOutter755Mode(); err != nil {
		fmt.Println("set path mode failed")
		return err
	}
	hwlog.RunLog.Info("-----set mef-center mode success-----")
	return nil
}

func (upf *UpgradePostFlowMgr) clearFlag() error {
	tgtPath, err := upf.InstallPathMgr.GetTargetWorkPath()
	if err != nil {
		hwlog.RunLog.Errorf("get backup work path failed: %s", err.Error())
		return errors.New("get backup work path failed")
	}

	flagPath := filepath.Join(tgtPath, util.UpgradeFlagFile)
	if err = fileutils.DeleteFile(flagPath); err != nil {
		hwlog.RunLog.Errorf("delete upgrade-flag failed: %s", err.Error())
		return errors.New("delete upgrade-flag failed")
	}
	return nil
}

func (upf *UpgradePostFlowMgr) clearEnv() {
	fmt.Println("upgrade failed, start to restore environment")
	hwlog.RunLog.Info("----------upgrade failed, start to restore environment-----------")
	clearMgr := util.GetUpgradeClearMgr(upf.SoftwareMgr, upf.step, upf.startedComponents, upf.notInstalledComponents)
	if err := clearMgr.ClearUpgrade(); err != nil {
		hwlog.RunLog.Errorf("clear upgrade environment failed: %s", err.Error())
		hwlog.RunLog.Error("----------upgrade failed, restore environment failed-----------")
		fmt.Println("clear environment failed, please recover it manually")
		return
	}
	fmt.Println("environment has been recovered")
	hwlog.RunLog.Info("----------upgrade failed, restore environment success-----------")
}
