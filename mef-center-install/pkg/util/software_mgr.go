// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindxedge/base/common"
)

// SoftwareMgr is the father struct of install/uninstall struct that manages functions to clear enviroment
type SoftwareMgr struct {
	Components     []string
	InstallPathMgr *InstallDirPathMgr
}

func (sm *SoftwareMgr) clearDockerImage() error {
	fmt.Println("start to clear docker image")
	hwlog.RunLog.Info("start to clear docker image")
	for _, name := range sm.Components {
		dockerMgr := GetDockerDealer(name, DockerTag)
		if err := dockerMgr.DeleteImage(); err != nil {
			return err
		}
	}
	fmt.Println("clear docker image success")
	hwlog.RunLog.Info("clear docker image success")
	return nil
}

func (sm *SoftwareMgr) clearInstallPkg() error {
	fmt.Println("start to clear install dir")
	hwlog.RunLog.Info("start to clear install dir")
	if !utils.IsExist(sm.InstallPathMgr.GetMefPath()) {
		hwlog.RunLog.Warn("mef-center install package does not exist, no need to delete")
		return nil
	}

	if err := common.DeleteAllFile(sm.InstallPathMgr.GetMefPath()); err != nil {
		hwlog.RunLog.Errorf("delete mef-center install package failed:%s", err.Error())
		return err
	}
	fmt.Println("clear install dir success")
	hwlog.RunLog.Info("clear install dir success")
	return nil
}

func (sm *SoftwareMgr) clearNodeLabel() error {
	fmt.Println("start to clear node label")
	hwlog.RunLog.Info("start to clear node label")
	localIp, err := GetLocalIp()
	if err != nil {
		hwlog.RunLog.Errorf("get local IP failed: %s", err.Error())
		return err
	}

	ipReg := fmt.Sprintf("'\\s%s\\s'", localIp)
	cmd := fmt.Sprintf(GetNodeCmdPattern, ipReg)
	nodeName, err := common.RunCommand("sh", false, "-c", cmd)
	if err != nil {
		hwlog.RunLog.Errorf("get current node failed: %s", err.Error())
		return err
	}

	// 删除不存在的label会显示执行命令成功
	_, err = common.RunCommand(CommandKubectl, true, "label", "node", nodeName, "mef-center-node-")
	if err != nil {
		hwlog.RunLog.Errorf("clear %s label command exec failed: %s", MefNamespace, err.Error())
		return errors.New("clear node label command exec failed")
	}
	fmt.Println("clear node label success")
	hwlog.RunLog.Info("clear node label success")
	return nil
}

// DoClear is the func that used to recover the environment that effected by installation
func (sm *SoftwareMgr) DoClear() error {
	var installTasks = []func() error{
		sm.clearDockerImage,
		sm.clearInstallPkg,
		sm.clearNodeLabel,
	}

	for _, function := range installTasks {
		if err := function(); err != nil {
			return err
		}
	}

	return nil
}
