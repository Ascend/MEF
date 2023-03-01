// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package control

import (
	"errors"
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
		sum.checkCurrentPath,
		sum.ClearNamespace,
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
		fmt.Println("current user is not root, cannot uninstall")
		hwlog.RunLog.Errorf("check user failed: %s", err.Error())
		return err
	}
	hwlog.RunLog.Info("check user successful")
	return nil
}

func (sum *SftUninstallMgr) checkCurrentPath() error {
	if err := util.CheckCurrentPath(sum.InstallPathMgr.GetWorkPath()); err != nil {
		fmt.Println("the existing dir is not the MEF working dir")
		hwlog.RunLog.Error(err)
		return errors.New("check current path failed")
	}
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
