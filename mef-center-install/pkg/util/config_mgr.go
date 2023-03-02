// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

// ConfigMgr is the struct to manager the mef-config dir
type ConfigMgr struct {
	cfgPathMgr *ConfigPathMgr
}

// GetConfigMgr is the func to init a ConfigMgr struct
func GetConfigMgr(pathMgr *ConfigPathMgr) *ConfigMgr {
	return &ConfigMgr{
		cfgPathMgr: pathMgr,
	}
}

// DoPrepare is the func to prepare mef-config dir on installation
func (cm *ConfigMgr) DoPrepare() error {
	var prepareConfigTasks = []func() error{
		cm.prepareConfigDir,
		cm.preparePubConfigDir,
		cm.prepareEdgeMgrFlag,
	}

	for _, function := range prepareConfigTasks {
		if err := function(); err != nil {
			return err
		}
	}

	return nil
}

func (cm *ConfigMgr) prepareConfigDir() error {
	configPath := cm.cfgPathMgr.GetConfigPath()
	if err := common.MakeSurePath(configPath); err != nil {
		hwlog.RunLog.Errorf("create config path [%s] failed: %v", configPath, err.Error())
		return errors.New("create config path failed")
	}
	return nil
}

func (cm *ConfigMgr) preparePubConfigDir() error {
	configPath := cm.cfgPathMgr.GetPublicConfigPath()
	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		hwlog.RunLog.Error("get current path failed")
		return errors.New("get current path failed")
	}
	installDir := filepath.Dir(currentDir)
	srcPath := filepath.Join(installDir, ConfigInPkg)
	dstPath := configPath

	if err = common.CopyDir(srcPath, dstPath, true); err != nil {
		hwlog.RunLog.Errorf("prepare public-config dir failed: %s", err.Error())
		return errors.New("prepare public-config dir failed")
	}

	return nil
}

func (cm *ConfigMgr) prepareEdgeMgrFlag() error {
	edgeMgrDirPath := cm.cfgPathMgr.GetComponentConfigPath(EdgeManagerName)
	if err := common.MakeSurePath(edgeMgrDirPath); err != nil {
		hwlog.RunLog.Errorf("prepare %s config dir failed: %v", EdgeManagerName, err.Error())
		return fmt.Errorf("prepare %s config dir failed", EdgeManagerName)
	}

	flagPath := cm.cfgPathMgr.GetEdgeMgrFlagPath()
	if err := common.CreateFile(flagPath, common.Mode600); err != nil {
		hwlog.RunLog.Errorf("create clear-flag path [%s] failed: %v", flagPath, err.Error())
		return errors.New("create clear-flag path failed")
	}
	return nil
}
