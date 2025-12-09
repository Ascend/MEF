// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package config this file for edge om config file manager
package config

import (
	"encoding/json"
	"errors"
	"fmt"

	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
)

// SmoothEdgeOmContainerConfig smooth edgeom container config to new edgeom config file
func SmoothEdgeOmContainerConfig(installRootDir string) error {
	edgeOmContainerConfigPath := pathmgr.NewConfigPathMgr(installRootDir).GetContainerConfigPath()
	edgeOmContainerConfig, err := util.LoadJsonFile(edgeOmContainerConfigPath)
	if err != nil {
		hwlog.RunLog.Errorf("get edgeom container config failed, error: %v", err)
		return errors.New("get edgeom container config failed")
	}

	maxContainerNumber := constants.MaxContainerNumber
	err = util.SetJsonValue(edgeOmContainerConfig, maxContainerNumber, constants.ConfigMaxContainerNumber)
	if err != nil {
		hwlog.RunLog.Errorf("set value for maxContainerNumber failed, error: %v", err)
		return errors.New("set value for maxContainerNumber failed")
	}

	var containerConfigData PodConfig
	arr, err := json.Marshal(edgeOmContainerConfig)
	if err != nil {
		hwlog.RunLog.Errorf("marshal edgeOmContainerConfig failed, error: %v", err)
		return errors.New("marshal edgeOmContainerConfig failed")
	}

	if err = json.Unmarshal(arr, &containerConfigData); err != nil {
		hwlog.RunLog.Errorf("unmarshal edgeOmContainerConfig failed, error: %v", err)
		return errors.New("unmarshal edgeOmContainerConfig failed")
	}

	addPaths := []string{"/usr/lib/aarch64-linux-gnu/libcrypto.so.1.1", "/usr/lib64/libcrypto.so.1.1",
		"/usr/lib/aarch64-linux-gnu/libyaml-0.so.2.0.6", "/var/lib/docker/modelfile"}
	newPaths := mergeArrayWithoutRepeat(containerConfigData.HostPath, addPaths)
	if err = util.SetJsonValue(edgeOmContainerConfig, newPaths, constants.ConfigHostPath); err != nil {
		hwlog.RunLog.Errorf("set value for hostPath failed, error: %v", err)
		return errors.New("set value for hostPath failed")
	}

	if err = fileutils.SetPathPermission(edgeOmContainerConfigPath, constants.Mode600,
		false, false); err != nil {
		hwlog.RunLog.Errorf("change edgeom container config path failed, error: %v", err)
		return errors.New("change edgeom container config path failed")
	}
	err1 := util.SaveJsonValue(edgeOmContainerConfigPath, edgeOmContainerConfig)

	if err = fileutils.SetPathPermission(edgeOmContainerConfigPath, constants.Mode400, false, false); err != nil {
		hwlog.RunLog.Errorf("post change edgeom container config path failed, error: %v", err)
		return errors.New("post change edgeom container config path failed")
	}
	if err1 != nil {
		hwlog.RunLog.Errorf("save edgeom container config failed, error: %v", err)
		return errors.New("save edgeom container config failed")
	}
	return nil
}

func isContain(items []string, item string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}

func mergeArrayWithoutRepeat(items1 []string, items2 []string) []string {
	mergedItems := items1
	for _, eachItem := range items2 {
		if isContain(items1, eachItem) {
			continue
		}
		mergedItems = append(mergedItems, eachItem)
	}
	return mergedItems
}

// SmoothAlarmConfigDB smooth alarm config db
func SmoothAlarmConfigDB() error {
	dbMgr, err := GetComponentDbMgr(constants.EdgeOm)
	if err != nil {
		return err
	}
	if err = dbMgr.InitDB(); err != nil {
		return errors.New("init alarm manager database failed")
	}

	if database.GetDb().Migrator().HasTable(AlarmConfig{}) {
		// table already exists, smooth fields.
		if err = database.CreateTableIfNotExist(AlarmConfig{}); err != nil {
			hwlog.RunLog.Errorf("migrate alarm config table failed, error: %v", err)
			return errors.New("migrate alarm config table failed")
		}
		hwlog.RunLog.Info("smooth alarm config success")
		return nil
	}

	if err = database.CreateTableIfNotExist(AlarmConfig{}); err != nil {
		hwlog.RunLog.Errorf("create alarm config table failed, error: %v", err)
		return errors.New("create alarm config table failed")
	}

	if err = SetDefaultAlarmCfg(); err != nil {
		hwlog.RunLog.Errorf("set default alarm config to table failed, error: %v", err)
		return errors.New("set default alarm config to table failed")
	}

	hwlog.RunLog.Info("smooth alarm config success")
	return nil
}

// SetDefaultAlarmCfg set default alarm config
func SetDefaultAlarmCfg() error {
	dbMgr, err := GetComponentDbMgr(constants.EdgeOm)
	if err != nil {
		return err
	}

	hasModified := util.GetBoolPointer(false)
	var alarmConfigs = []AlarmConfig{
		{constants.CertCheckPeriodDB, constants.DefaultCheckPeriod, hasModified},
		{constants.CertOverdueThresholdDB, constants.DefaultOverdueThreshold, hasModified},
	}

	for _, cfg := range alarmConfigs {
		if err = dbMgr.SetAlarmConfig(&cfg); err != nil {
			hwlog.RunLog.Errorf("set alarm config %s failed, error: %v", cfg.ConfigName, err)
			return fmt.Errorf("set alarm config %s failed", cfg.ConfigName)
		}
	}

	hwlog.RunLog.Info("set default alarm config success")
	return nil
}
