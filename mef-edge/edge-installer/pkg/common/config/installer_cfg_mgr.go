// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package config this file for installer config file manager
package config

import (
	"errors"
	"strings"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
)

// SetNetManager set net manager config
func SetNetManager(dbMgr *DbMgr, netMgr *NetManager) error {
	if netMgr == nil {
		return errors.New("set net manager failed, input is nil")
	}
	return dbMgr.SetConfig(constants.NetMgrConfigKey, netMgr)
}

// GetNetManager get net manager config
func GetNetManager(dbMgr *DbMgr) (*NetManager, error) {
	var netConfig NetManager
	if err := dbMgr.GetConfig(constants.NetMgrConfigKey, &netConfig); err != nil {
		return nil, err
	}
	return &netConfig, nil
}

// GetNodeId get node id
func GetNodeId(dbMgr *DbMgr) string {
	installerConfig, err := GetInstall(dbMgr)
	if err != nil {
		hwlog.RunLog.Errorf("get net manager config failed, error: %v", err)
		return ""
	}
	return installerConfig.SerialNumber
}

// SetInstall set install config
func SetInstall(dbMgr *DbMgr, installCfg *InstallerConfig) error {
	return dbMgr.SetConfig(constants.InstallerConfigKey, installCfg)
}

// GetInstall get install config
func GetInstall(dbMgr *DbMgr) (*InstallerConfig, error) {
	var installCfg InstallerConfig
	if err := dbMgr.GetConfig(constants.InstallerConfigKey, &installCfg); err != nil {
		return nil, err
	}
	return &installCfg, nil
}

// CheckIsA500 check device is a500
func CheckIsA500() bool {
	output, err := envutils.RunCommand(constants.NpuSmiCmd, envutils.DefCmdTimeoutSec,
		"info", "-t", "product", "-i", "0")
	if err == nil && (strings.Contains(output, constants.A500Name) ||
		strings.Contains(output, constants.A500NameOld)) {
		return true
	}
	return false
}
