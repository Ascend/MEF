// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
)

// ConfigMgr is the struct to manager the mef-config dir
type ConfigMgr struct {
	cfgPathMgr *ConfigPathMgr
	components []string
}

// GetConfigMgr is the func to init a ConfigMgr struct
func GetConfigMgr(pathMgr *ConfigPathMgr, components []string) *ConfigMgr {
	return &ConfigMgr{
		cfgPathMgr: pathMgr,
		components: components,
	}
}

// DoPrepare is the func to prepare mef-config dir on installation
func (cm *ConfigMgr) DoPrepare() error {
	var prepareConfigTasks = []func() error{
		cm.prepareConfigDir,
		cm.preparePubConfigDir,
		cm.prepareComponentsConfig,
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

func (cm *ConfigMgr) prepareComponentsConfig() error {
	for _, component := range cm.components {
		componentMgr := GetComponentMgr(component)
		if err := componentMgr.PrepareComponentConfig(cm.cfgPathMgr); err != nil {
			return err
		}
	}

	return nil
}

// InitAndSetAlarmCfgTable init and set alarm config
func InitAndSetAlarmCfgTable(configDir string) error {
	if err := database.CreateTableIfNotExist(common.AlarmConfig{}); err != nil {
		hwlog.RunLog.Errorf("create alarm config table failed, error: %v", err)
		return errors.New("create alarm config table failed")
	}

	dbMgr := common.NewDbMgr(configDir, AlarmConfigDB)
	hasModified := GetBoolPointer(false)
	var alarmConfigs = []common.AlarmConfig{
		{CertCheckPeriodDB, DefaultCheckPeriod, hasModified},
		{CertOverdueThresholdDB, DefaultOverdueThreshold, hasModified},
	}

	for _, cfg := range alarmConfigs {
		if err := dbMgr.SetAlarmConfig(&cfg); err != nil {
			hwlog.RunLog.Errorf("set alarm config %s failed, error: %v", cfg.ConfigName, err)
			return fmt.Errorf("set alarm config failed, error: %v", err)
		}
	}

	hwlog.RunLog.Info("set alarm config success")
	return nil
}
