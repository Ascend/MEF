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

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/install"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// UpgradePostFlowMgr is a struct that used to uninstall mef-center
type UpgradePostFlowMgr struct {
	util.SoftwareMgr
	logPathMgr        *util.LogDirPathMgr
	startedComponents []string
	step              int
}

// GetUpgradePostMgr is a func to init an UpgradePostFlowMgr struct
func GetUpgradePostMgr(components []string, installPath string,
	logRootPath string, logBackupRootPath string) *UpgradePostFlowMgr {
	return &UpgradePostFlowMgr{
		SoftwareMgr: util.SoftwareMgr{
			Components:     components,
			InstallPathMgr: util.InitInstallDirPathMgr(installPath),
		},
		logPathMgr: util.InitLogDirPathMgr(logRootPath, logBackupRootPath),
		step:       util.ClearUnpackPathStep,
	}
}

// DoUpgrade is the main flow-control func to exec upgrade flow on the new package
func (upf *UpgradePostFlowMgr) DoUpgrade() error {
	var installTasks = []func() error{
		upf.checkVersion,
		upf.checkNecessaryTools,
		upf.prepareK8sLabel,
		upf.createFlag,
		upf.prepareWorkCDir,
		upf.prepareYaml,
		upf.recordStarted,
		upf.deleteNameSpace,
		upf.removeDockerImage,
		upf.buildNewImage,
		upf.startNewPod,
		upf.resetSoftLink,
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
	if err := common.CreateFile(upf.InstallPathMgr.WorkPathMgr.GetUpgradeFlagPath(), common.Mode400); err != nil {
		hwlog.RunLog.Errorf("create upgrade-flag failed: %s", err.Error())
		return errors.New("create upgrade-flag failed")
	}
	return nil
}

func (upf *UpgradePostFlowMgr) prepareWorkCDir() error {
	upf.step = util.ClearTempUpgradePathStep
	if err := common.MakeSurePath(upf.InstallPathMgr.TmpPathMgr.GetWorkPath()); err != nil {
		hwlog.RunLog.Errorf("make sure temp upgrade dir failed: %s", err.Error())
		return errors.New("make sure temp upgrade dir failed")
	}

	workingDirHandleCtl := install.GetWorkingDirMgr(upf.InstallPathMgr.TmpPathMgr,
		upf.InstallPathMgr.GetWorkPath(), upf.Components)

	if err := workingDirHandleCtl.DoUpgradePrepare(); err != nil {
		hwlog.RunLog.Errorf("prepare working dir failed: %v", err.Error())
		return errors.New("prepare working dir failed")
	}

	if err := common.DeleteAllFile(upf.InstallPathMgr.WorkPathMgr.GetRelativeVarDirPath()); err != nil {
		hwlog.RunLog.Errorf("delete unpack dir failed: %s", err.Error())
		return errors.New("delete unpack dir failed")
	}
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
	hwlog.RunLog.Info("prepare components' yaml successful")
	return nil
}

func (upf *UpgradePostFlowMgr) recordStarted() error {
	hwlog.RunLog.Info("start to record started components")
	for _, c := range upf.SoftwareMgr.Components {
		dealer := &util.CtlComponent{
			Name:           c,
			InstallPathMgr: upf.SoftwareMgr.InstallPathMgr,
		}
		started, err := dealer.CheckStarted()
		if err != nil {
			hwlog.RunLog.Errorf("check component %s's status failed", c)
			return fmt.Errorf("check component %s's status failed", c)
		}

		if !started {
			continue
		}
		hwlog.RunLog.Infof("component %s is running", c)
		upf.startedComponents = append(upf.startedComponents, c)
	}
	hwlog.RunLog.Info("record started components succeeds")
	return nil
}

func (upf *UpgradePostFlowMgr) deleteNameSpace() error {
	upf.step = util.RestartPodStep
	hwlog.RunLog.Info("start to delete mef-center namespace")
	namespaceMgr := util.NewNamespaceMgr(util.MefNamespace)
	if err := namespaceMgr.ClearNamespace(); err != nil {
		return err
	}
	hwlog.RunLog.Info("delete mef-center namespace succeeds")
	return nil
}

func (upf *UpgradePostFlowMgr) removeDockerImage() error {
	upf.step = util.LoadOldDockerStep
	for _, component := range upf.SoftwareMgr.Components {
		dockerDealerIns := util.GetDockerDealer(component, util.DockerTag)
		if err := dockerDealerIns.DeleteImage(); err != nil {
			return err
		}
	}

	return nil
}

func (upf *UpgradePostFlowMgr) buildNewImage() error {
	hwlog.RunLog.Info("start to build new docker image")
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
	return nil
}

func (upf *UpgradePostFlowMgr) startNewPod() error {
	upf.step = util.ClearNameSpaceStep
	for _, c := range upf.startedComponents {
		dealer := &util.CtlComponent{
			Name:           c,
			Operation:      util.StartOperateFlag,
			InstallPathMgr: upf.SoftwareMgr.InstallPathMgr,
		}
		err := dealer.Operate()
		if err != nil {
			hwlog.RunLog.Errorf("start component %s failed", c)
			return fmt.Errorf("start component %s failed", c)
		}
	}
	return nil
}

func (upf *UpgradePostFlowMgr) resetSoftLink() error {
	backPath, err := upf.InstallPathMgr.GetTargetWorkPath()
	if err != nil {
		hwlog.RunLog.Errorf("get target work path failed: %s", err.Error())
		return err
	}

	if err = common.DeleteAllFile(backPath); err != nil {
		hwlog.RunLog.Errorf("delete backUp path failed: %s", err.Error())
		return err
	}

	if err = common.RenameFile(upf.InstallPathMgr.GetTmpUpgradePath(), backPath); err != nil {
		hwlog.RunLog.Errorf("rename temp-upgrade dir failed: %s", err.Error())
		return err
	}

	if utils.IsExist(upf.InstallPathMgr.GetWorkPath()) {
		if err = os.Remove(upf.InstallPathMgr.GetWorkPath()); err != nil {
			hwlog.RunLog.Errorf("remove old software dir symlink failed, error: %s", err.Error())
			return errors.New("remove old software dir symlink failed")
		}
	}

	if err = common.CreateSoftLink(backPath, upf.InstallPathMgr.GetWorkPath()); err != nil {
		hwlog.RunLog.Errorf("create software dir symlink failed, error: %s", err.Error())
		return errors.New("create software dir symlink failed")
	}

	return nil
}

func (upf *UpgradePostFlowMgr) clearFlag() error {
	tgtPath, err := upf.InstallPathMgr.GetTargetWorkPath()
	if err != nil {
		hwlog.RunLog.Errorf("get backup work path failed: %s", err.Error())
		return errors.New("get backup work path failed")
	}

	flagPath := filepath.Join(tgtPath, util.UpgradeFlagFile)
	if err = common.DeleteFile(flagPath); err != nil {
		hwlog.RunLog.Errorf("delete upgrade-flag failed: %s", err.Error())
		return errors.New("delete upgrade-flag failed")
	}
	return nil
}

func (upf *UpgradePostFlowMgr) clearEnv() {
	fmt.Println("upgrade failed, start to restore environment")
	hwlog.RunLog.Info("----------upgrade failed, start to restore environment-----------")
	clearMgr := util.GetUpgradeClearMgr(upf.SoftwareMgr, upf.step, upf.startedComponents)
	if err := clearMgr.ClearUpgrade(); err != nil {
		hwlog.RunLog.Errorf("clear upgrade environment failed: %s", err.Error())
		hwlog.RunLog.Error("----------upgrade failed, restore environment failed-----------")
		fmt.Println("clear environment failed, plz recover it manually")
		return
	}
	fmt.Println("environment has been recovered")
	hwlog.RunLog.Info("----------upgrade failed, restore environment success-----------")
}
