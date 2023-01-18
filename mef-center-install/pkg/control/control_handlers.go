// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package control contains the unique method for start/stop/restart component
package control

import (
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// SftControlMgr is a struct that used to start/stop/restart a component
type SftControlMgr struct {
	componentFlag      string
	operate            string
	installedComponent []string
	installPathMgr     *util.InstallDirPathMgr
	componentList      []*util.CtlComponent
}

// DoControl is the main func to do a control handle
func (scm *SftControlMgr) DoControl() error {
	var installTasks = []func() error{
		scm.init,
		scm.deal,
	}

	for _, function := range installTasks {
		if err := function(); err != nil {
			return err
		}
	}

	return nil
}

func (scm *SftControlMgr) init() error {
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

func (scm *SftControlMgr) deal() error {
	for _, component := range scm.componentList {
		if err := component.Operate(); err != nil {
			return err
		}
	}
	return nil
}

// InitSftControlMgr is used to init a SftControlMgr struct
func InitSftControlMgr(component, operate string,
	installComponents []string, installPathMgr *util.InstallDirPathMgr) *SftControlMgr {
	return &SftControlMgr{
		componentFlag:      component,
		operate:            operate,
		installedComponent: installComponents,
		installPathMgr:     installPathMgr,
	}
}
