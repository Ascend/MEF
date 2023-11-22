// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

package util

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
)

// UpgradeClearMgr is the mgr to restore environment for upgrade flow
type UpgradeClearMgr struct {
	SoftwareMgr
	step               int
	startedComponent   []string
	installedComponent []string
}

// GetUpgradeClearMgr is the func to init an UpgradeClearMgr mgt
func GetUpgradeClearMgr(softwareMgr SoftwareMgr, step int, startedComponent []string,
	notInstalledComponent []string) *UpgradeClearMgr {
	allSet := utils.NewSet(softwareMgr.Components...)
	startedSet := utils.NewSet(startedComponent...)
	notInstalledSet := utils.NewSet(notInstalledComponent...)

	return &UpgradeClearMgr{
		SoftwareMgr:        softwareMgr,
		step:               step,
		startedComponent:   startedSet.Difference(notInstalledSet).List(),
		installedComponent: allSet.Difference(notInstalledSet).List(),
	}
}

func (ucm *UpgradeClearMgr) reloadDockerImage() error {
	fmt.Println("start to reload docker image")
	hwlog.RunLog.Info("start to reload docker image")
	for _, component := range ucm.installedComponent {
		dockerDealerIns := GetAscendDockerDealer(component, DockerTag)
		imagePath := ucm.InstallPathMgr.WorkPathMgr.GetImageDirPath(component)
		if err := dockerDealerIns.ReloadImage(imagePath); err != nil {
			fmt.Println("reload docker image failed")
			hwlog.RunLog.Errorf("reload component %s's docker failed:%s", component, err.Error())
			return fmt.Errorf("reload component %s's docker failed", component)
		}
	}
	fmt.Println("reload docker image success")
	hwlog.RunLog.Info("reload docker image success")

	return nil
}

func (ucm *UpgradeClearMgr) restartPods() error {
	fmt.Println("start to restart pods")
	hwlog.RunLog.Info("start to restart pods")
	nsMgr := NewNamespaceMgr(MefNamespace)
	if err := nsMgr.prepareNameSpace(); err != nil {
		fmt.Println("restart pods failed: prepare namespace failed")
		hwlog.RunLog.Errorf("prepare %s namespace failed: %s", MefNamespace, err.Error())
		return fmt.Errorf("prepare %s namespace failed", MefNamespace)
	}

	for _, component := range ucm.startedComponent {
		componentMgr := &CtlComponent{
			Name:           component,
			Operation:      StartOperateFlag,
			InstallPathMgr: ucm.InstallPathMgr.WorkPathMgr,
		}

		if err := componentMgr.Operate(); err != nil {
			fmt.Println("restart pods failed: restart pods failed")
			return err
		}
	}

	fmt.Println("restart pods success")
	hwlog.RunLog.Info("restart pods success ")
	return nil
}

func (ucm *UpgradeClearMgr) clearTempUpgradePath() error {
	fmt.Println("start to clear temp upgrade path")
	hwlog.RunLog.Info("start to clear temp upgrade path")
	if err := fileutils.DeleteAllFileWithConfusion(ucm.InstallPathMgr.TmpPathMgr.GetWorkPath()); err != nil {
		fmt.Println("clear temp-upgrade dir failed")
		hwlog.RunLog.Errorf("clear temp-upgrade dir failed: %s", err.Error())
		return fmt.Errorf("clear temp-upgrade dir failed")
	}
	fmt.Println("clear temp upgrade path success")
	hwlog.RunLog.Info("clear temp upgrade path success")
	return nil
}

func (ucm *UpgradeClearMgr) clearUnpackPath() error {
	fmt.Println("start to clear unpack path")
	hwlog.RunLog.Info("start to clear unpack path")
	unpackPath := ucm.InstallPathMgr.WorkPathMgr.GetVarDirPath()
	if fileutils.IsExist(unpackPath) {
		if err := fileutils.DeleteAllFileWithConfusion(unpackPath); err != nil {
			fmt.Println("clear unpack path failed")
			hwlog.RunLog.Errorf("clear unpack path failed: %s", err.Error())
			return fmt.Errorf("clear unpack path failed")
		}
	}
	fmt.Println("clear unpack path success")
	hwlog.RunLog.Info("clear unpack path success")
	return nil
}

// ClearUpgrade is the func that used to clear the environment when upgrade failed
func (ucm *UpgradeClearMgr) ClearUpgrade() error {
	if ucm.step >= ClearNameSpaceStep {
		if err := ucm.ClearMEFCenterNamespace(); err != nil {
			hwlog.RunLog.Errorf("delete mef center namespace failed: %s", err.Error())
			return err
		}
	}

	if ucm.step >= RemoveDockerStep {
		if err := ucm.ClearAllDockerImages(); err != nil {
			hwlog.RunLog.Errorf("clear Docker Image failed: %s", err.Error())
			return err
		}
	}

	if ucm.step >= LoadOldDockerStep {
		if err := ucm.reloadDockerImage(); err != nil {
			return err
		}
	}

	if ucm.step >= RestartPodStep {
		if err := ucm.restartPods(); err != nil {
			return err
		}
	}

	if ucm.step >= ClearTempUpgradePathStep {
		if err := ucm.clearTempUpgradePath(); err != nil {
			return err
		}
	}

	if err := ucm.clearUnpackPath(); err != nil {
		return err
	}

	flagPath := ucm.InstallPathMgr.WorkPathMgr.GetUpgradeFlagPath()
	if !fileutils.IsExist(flagPath) {
		return nil
	}
	if err := fileutils.DeleteFile(flagPath); err != nil {
		hwlog.RunLog.Errorf("delete upgrade-flag failed: %s", err.Error())
		return errors.New("delete upgrade-flag failed")
	}

	return nil
}
