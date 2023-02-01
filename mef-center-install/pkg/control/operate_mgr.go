// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package control contains the unique method for start/stop/restart component
package control

import (
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// SftOperateMgr is a struct that used to start/stop/restart a component
type SftOperateMgr struct {
	componentFlag      string
	operate            string
	installedComponent []string
	installPathMgr     *util.InstallDirPathMgr
	componentList      []*util.CtlComponent
}

// DoOperate is the main func to do an operate handle
func (scm *SftOperateMgr) DoOperate() error {
	var controlTasks = []func() error{
		scm.init,
		scm.check,
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
				InstallPathMgr: scm.installPathMgr,
			}
			scm.componentList = append(scm.componentList, component)
		}
	} else {
		// if just a certain componentFlag, then construct a single-element componentFlag list
		component := &util.CtlComponent{
			Name:           scm.componentFlag,
			Operation:      scm.operate,
			InstallPathMgr: scm.installPathMgr,
		}
		scm.componentList = append(scm.componentList, component)
	}

	hwlog.RunLog.Info("init componentFlag list successful")
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
	if err := util.CheckUser(); err != nil {
		hwlog.RunLog.Errorf("check user failed: %s", err.Error())
		return err
	}
	hwlog.RunLog.Info("check user successful")
	return nil
}

func (scm *SftOperateMgr) deal() error {
	for _, component := range scm.componentList {
		if err := component.Operate(); err != nil {
			return err
		}
	}
	return nil
}

// InitSftOperateMgr is used to init a SftOperateMgr struct
func InitSftOperateMgr(component, operate string,
	installComponents []string, installPath string) *SftOperateMgr {
	return &SftOperateMgr{
		componentFlag:      component,
		operate:            operate,
		installedComponent: installComponents,
		installPathMgr:     util.InitInstallDirPathMgr(installPath),
	}
}
