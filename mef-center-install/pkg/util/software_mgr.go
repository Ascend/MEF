// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"errors"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
)

// SoftwareMgr is the father struct of install/uninstall/upgrade struct that manages public functions
type SoftwareMgr struct {
	Components     []string
	InstallPathMgr *InstallDirPathMgr
}

// ClearDockerImage is used to clear docker images
func (sm *SoftwareMgr) ClearDockerImage(components []string) error {
	fmt.Println("start to clear docker image")
	hwlog.RunLog.Info("start to clear docker image")
	for _, name := range components {
		dockerMgr := GetDockerDealer(name, DockerTag)
		if err := dockerMgr.DeleteImage(); err != nil {
			hwlog.RunLog.Errorf("delete %s image failed, error: %v", name, err)
			return err
		}
	}
	fmt.Println("clear docker image success")
	hwlog.RunLog.Info("clear docker image success")
	return nil
}

// ClearAllDockerImages is used to clear all docker images for installed components
func (sm *SoftwareMgr) ClearAllDockerImages() error {
	return sm.ClearDockerImage(sm.Components)
}

func (sm *SoftwareMgr) clearInstallPkg() error {
	fmt.Println("start to clear install dir")
	hwlog.RunLog.Info("start to clear install dir")
	if !fileutils.IsExist(sm.InstallPathMgr.GetMefPath()) {
		hwlog.RunLog.Warn("mef-center install package does not exist, no need to delete")
		return nil
	}
	hwlog.RunLog.Infof("All key in %s are about to be destroyed.", sm.InstallPathMgr.GetMefPath())
	if err := fileutils.DeleteAllFileWithConfusion(sm.InstallPathMgr.GetMefPath()); err != nil {
		hwlog.RunLog.Errorf("delete mef-center install package failed:%s", err.Error())
		return err
	}

	fmt.Println("clear install dir success")
	hwlog.RunLog.Info("clear install dir success")
	return nil
}

func (sm *SoftwareMgr) clearLock() error {
	fmt.Println("start to clear lock")
	hwlog.RunLog.Info("start to clear lock")
	lockPath := filepath.Join("/run", MefCenterLock)

	if err := fileutils.DeleteFile(lockPath); err != nil {
		hwlog.RunLog.Warnf("delete mef-center lock failed: %s", err.Error())
		fmt.Println("warning: clear mef-center lock failed")
		return nil
	}

	fmt.Println("clear mef-center lock success")
	hwlog.RunLog.Info("clear mef-center lock success")
	return nil
}

func (sm *SoftwareMgr) clearNodeLabel() error {
	fmt.Println("start to clear node label")
	hwlog.RunLog.Info("start to clear node label")
	var k8slMgr = K8sLabelMgr{}
	nodeName, err := k8slMgr.getMasterNodeName()
	if err != nil {
		return err
	}

	// 删除不存在的label会显示执行命令成功
	_, err = envutils.RunCommand(CommandKubectl, envutils.DefCmdTimeoutSec,
		"label", "node", nodeName, "mef-center-node-")
	if err != nil {
		hwlog.RunLog.Errorf("clear %s label command exec failed: %s", MefNamespace, err.Error())
		return errors.New("clear node label command exec failed")
	}
	fmt.Println("clear node label success")
	hwlog.RunLog.Info("clear node label success")
	return nil
}

// ClearMEFUserNamespace is used to clear mef-user namespace
func (sm *SoftwareMgr) ClearMEFUserNamespace() error {
	fmt.Println("start to clear Namespace")
	hwlog.RunLog.Info("start to clear Namespace")
	// mef-users namespace creates by edge-manager
	nsMgr := NewNamespaceMgr(common.MefUserNs)
	if err := nsMgr.ForceClearNamespace(); err != nil {
		fmt.Printf("clear %s namespace failed\n", common.MefUserNs)
		hwlog.RunLog.Errorf("clear %s namespace failed\n", common.MefUserNs)
		return err
	}

	fmt.Printf("clear Namespace[%s] success\n", common.MefUserNs)
	hwlog.RunLog.Infof("clear Namespace[%s] success", common.MefUserNs)
	return nil
}

// ClearMEFCenterNamespace delete mef-center namespace
func (sm *SoftwareMgr) ClearMEFCenterNamespace() error {
	fmt.Printf("start to clear Namespace[%s]\n", MefNamespace)
	hwlog.RunLog.Infof("start to clear Namespace[%s]", MefNamespace)

	nsMgr := NewNamespaceMgr(MefNamespace)
	if err := nsMgr.ClearNamespace(); err != nil {
		fmt.Printf("clear %s namespace failed\n", MefNamespace)
		hwlog.RunLog.Errorf("clear %s namespace failed\n", MefNamespace)
		return err
	}

	fmt.Printf("clear Namespace success[%s]\n", MefNamespace)
	hwlog.RunLog.Infof("clear Namespace success[%s]", MefNamespace)
	return nil
}

// ClearKubeAuth is used to clear mef-center k8s auth
func (sm *SoftwareMgr) ClearKubeAuth() error {
	fmt.Println("start to clear K8s auth")
	hwlog.RunLog.Info("start to clear K8s auth")

	var err error
	if _, err = envutils.RunCommand(CommandKubectl, envutils.DefCmdTimeoutSec, "delete",
		"clusterrole", clusterroleName); err != nil {
		hwlog.RunLog.Warnf("delete clusterrole failed: %v", err)
	}
	if _, err = envutils.RunCommand(CommandKubectl, envutils.DefCmdTimeoutSec, "delete",
		"clusterrolebinding", bindingName); err != nil {
		hwlog.RunLog.Warnf("delete clusterrolebinding failed: %v", err)
	}
	if err != nil {
		fmt.Println("warning: clear K8s auth failed")
		return nil
	}
	fmt.Println("clear K8s auth success")
	hwlog.RunLog.Info("clear K8s auth success")
	return nil
}

// ClearAndLabel is the func that used to recover the environment that effected by installation
func (sm *SoftwareMgr) ClearAndLabel() error {
	var installTasks = []func() error{
		sm.clearNodeLabel,
		sm.clearInstallPkg,
		sm.clearLock,
	}

	for _, function := range installTasks {
		if err := function(); err != nil {
			return err
		}
	}

	return nil
}
