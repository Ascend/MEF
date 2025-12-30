// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package control

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/envutils"
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
		sum.ClearMEFCenterNamespace,
		sum.ClearMEFUserNamespace,
		sum.ClearKubeAuth,
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
	if err := envutils.CheckUserIsRoot(); err != nil {
		fmt.Println("current user is not root, cannot uninstall")
		hwlog.RunLog.Errorf("check user failed: %s", err.Error())
		return errors.New("check user failed")
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
func GetSftUninstallMgrIns(components []string, installPathMgr *util.InstallDirPathMgr) SftUninstallMgr {
	return SftUninstallMgr{
		SoftwareMgr: util.SoftwareMgr{
			Components:     components,
			InstallPathMgr: installPathMgr,
		},
	}
}
