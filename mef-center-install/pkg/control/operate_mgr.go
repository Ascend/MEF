// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package control contains the unique method for start/stop/restart component
package control

import (
	"errors"
	"fmt"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// SftOperateMgr is a struct that used to start/stop/restart a component
type SftOperateMgr struct {
	componentFlag      string
	operate            string
	installedComponent []string
	installPathMgr     *util.InstallDirPathMgr
	logPathMgr         *util.LogDirPathMgr
	componentList      []*util.CtlComponent
	optionComList      []util.OptionComponent
}

// DoOperate is the main func to do an Operate handle
func (scm *SftOperateMgr) DoOperate() error {
	var controlTasks = []func() error{
		scm.init,
		scm.prepareComponentLogDir,
		scm.prepareComponentLogBackupDir,
		scm.check,
		scm.dealUpgradeFlag,
		scm.deal,
	}

	for _, function := range controlTasks {
		if err := function(); err != nil {
			return err
		}
	}

	return nil
}

func (scm *SftOperateMgr) init() error {
	// if all, then construct a full componentFlag list. (Does not support batch configuration)
	if scm.componentFlag == "all" {
		for _, c := range scm.installedComponent {
			component := &util.CtlComponent{
				Name:           c,
				Operation:      scm.operate,
				InstallPathMgr: scm.installPathMgr.WorkPathMgr,
			}
			scm.componentList = append(scm.componentList, component)
		}
		installInfo, err := util.GetInstallInfo()
		if err != nil {
			return err
		}
		for _, c := range installInfo.OptionComponent {
			component := util.OptionComponent{
				Name:      c,
				Operation: scm.operate,
				PathMgr:   scm.installPathMgr,
			}
			scm.optionComList = append(scm.optionComList, component)
		}
	} else {
		// if just a certain componentFlag, then construct a single-element componentFlag list
		if err := util.CheckParamOption(util.OptionalComponent(), scm.componentFlag); err == nil {
			component := util.OptionComponent{
				Name:      scm.componentFlag,
				Operation: scm.operate,
				PathMgr:   scm.installPathMgr,
			}
			scm.optionComList = append(scm.optionComList, component)
		} else {
			component := &util.CtlComponent{
				Name:           scm.componentFlag,
				Operation:      scm.operate,
				InstallPathMgr: scm.installPathMgr.WorkPathMgr,
			}
			scm.componentList = append(scm.componentList, component)
		}
	}

	hwlog.RunLog.Info("init componentFlag list successful")
	return nil
}

func (scm *SftOperateMgr) prepareComponentLogDir() error {
	hwlog.RunLog.Info("start to prepare components' log dir")
	for _, component := range scm.installedComponent {
		componentMgr := util.GetComponentMgr(component)
		if err := componentMgr.PrepareLogDir(scm.logPathMgr); err != nil {
			return err
		}
	}
	hwlog.RunLog.Info("prepare components' log dir successful")
	return nil
}

func (scm *SftOperateMgr) prepareComponentLogBackupDir() error {
	hwlog.RunLog.Info("start to prepare components' log backup dir")
	for _, component := range scm.installedComponent {
		componentMgr := util.GetComponentMgr(component)
		if err := componentMgr.PrepareLogBackupDir(scm.logPathMgr); err != nil {
			return err
		}
	}
	hwlog.RunLog.Info("prepare components' log backup dir successful")
	return nil
}

func (scm *SftOperateMgr) check() error {
	var checkTasks = []func() error{
		scm.checkUser,
	}

	for _, function := range checkTasks {
		if err := function(); err != nil {
			return err
		}
	}

	return nil
}

func (scm *SftOperateMgr) checkUser() error {
	hwlog.RunLog.Info("start to check user")
	if err := envutils.CheckUserIsRoot(); err != nil {
		fmt.Println("the current user is not root, cannot operate MEF-Center")
		hwlog.RunLog.Errorf("check user failed: %s", err.Error())
		return err
	}
	hwlog.RunLog.Info("check user successful")
	return nil
}

func (scm *SftOperateMgr) dealUpgradeFlag() error {
	if !utils.IsExist(scm.installPathMgr.WorkPathMgr.GetUpgradeFlagPath()) {
		return nil
	}

	if scm.operate != util.StartOperateFlag {
		fmt.Println("the last upgrade was terminated unexpectedly, plz use start operate to recovery first")
		hwlog.RunLog.Error("the last upgrade was terminated unexpectedly, needs to recovery first")
		return errors.New("needs to recovery environment")
	}

	fmt.Println("find upgrade flag exist, start to recover environment")
	hwlog.RunLog.Info("----------find upgrade flag exist, start to recover environment-----------")
	clearMgr := util.GetUpgradeClearMgr(util.SoftwareMgr{
		Components:     scm.installedComponent,
		InstallPathMgr: scm.installPathMgr,
	}, util.ClearNameSpaceStep, []string{}, []string{})
	if err := clearMgr.ClearUpgrade(); err != nil {
		hwlog.RunLog.Errorf("clear upgrade environment failed: %s", err.Error())
		hwlog.RunLog.Error("----------restore environment failed-----------")
		fmt.Println("clear environment failed, plz recover it manually")
		return err
	}
	fmt.Println("environment has been recovered")
	hwlog.RunLog.Info("----------restore environment success-----------")
	return nil
}

func (scm *SftOperateMgr) deal() error {
	var failedList []string
	for _, c := range scm.optionComList {
		if c.Name == util.IcsManagerName {
			ics := icsManager{pathMgr: scm.installPathMgr, name: util.IcsManagerName, operate: scm.operate}
			if err := ics.Operate(); err != nil {
				fmt.Printf("%s component %s failed\n", scm.operate, ics.name)
				failedList = append(failedList, c.Name)
			}
		}
	}
	for _, component := range scm.componentList {
		if err := component.Operate(); err != nil {
			fmt.Printf("%s component %s failed\n", component.Operation, component.Name)
			failedList = append(failedList, component.Name)
		}
	}
	if len(failedList) != 0 {
		fmt.Printf("%s operation on components %s failed\n", scm.operate, failedList)
		hwlog.RunLog.Errorf("%s operation on components %s failed", scm.operate, failedList)
		return fmt.Errorf("%s operation on components %s failed", scm.operate, failedList)
	}
	return nil
}

// InitSftOperateMgr is used to init a SftOperateMgr struct
func InitSftOperateMgr(component, operate string,
	installComponents []string, installPathMgr *util.InstallDirPathMgr, logPathMgr *util.LogDirPathMgr) *SftOperateMgr {
	return &SftOperateMgr{
		componentFlag:      component,
		operate:            operate,
		installedComponent: installComponents,
		installPathMgr:     installPathMgr,
		logPathMgr:         logPathMgr,
	}
}
