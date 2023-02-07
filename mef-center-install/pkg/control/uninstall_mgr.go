// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package control

import (
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// SftUninstallMgr is a struct that used to uninstall mef-center
type SftUninstallMgr struct {
	util.SoftwareMgr
}

// DoUninstall is the main func that to uninstall mef-center
func (sum *SftUninstallMgr) DoUninstall() error {
	var installTasks = []func() error{
		sum.checkUser,
		sum.clearNamespace,
		sum.ClearAllDockerImages,
		sum.ClearAndLabel,
	}

	for _, function := range installTasks {
		if err := function(); err != nil {
			return err
		}
	}

	return nil
}

func (sum *SftUninstallMgr) checkUser() error {
	hwlog.RunLog.Info("start to check user")
	if err := util.CheckUser(); err != nil {
		hwlog.RunLog.Errorf("check user failed: %s", err.Error())
		return err
	}
	hwlog.RunLog.Info("check user successful")
	return nil
}

func (sum *SftUninstallMgr) clearNamespace() error {
	fmt.Println("start to clear Namespace")
	hwlog.RunLog.Info("start to clear Namespace")
	nsMgr := util.NewNamespaceMgr(util.MefNamespace)
	if err := nsMgr.ClearNamespace(); err != nil {
		fmt.Println("clear Namespace failed")
		return err
	}
	fmt.Println("clear Namespace success")
	hwlog.RunLog.Info("clear Namespace success")
	return nil
}

// GetSftUninstallMgrIns is used to init a SftUninstallMgrIns struct
func GetSftUninstallMgrIns(components []string, installPath string) SftUninstallMgr {
	return SftUninstallMgr{
		SoftwareMgr: util.SoftwareMgr{
			Components:     components,
			InstallPathMgr: util.InitInstallDirPathMgr(installPath),
		},
	}
}
