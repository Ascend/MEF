// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tasks for setting system info task
package tasks

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
)

// SetSystemInfoTask the task for set system info
type SetSystemInfoTask struct {
	ConfigDir     string
	ConfigPathMgr *pathmgr.ConfigPathMgr
	LogPathMgr    *pathmgr.LogPathMgr
	dbMgr         *config.DbMgr
}

// Run set system info task
func (ssi *SetSystemInfoTask) Run() error {
	var setFunc = []func() error{
		ssi.prepareEdgeCoreDb,
		ssi.initOmDb,
		ssi.setInstallConfig,
		ssi.setEdgeCoreConfig,
		ssi.setDefaultNetConfig,
		ssi.setDefaultAlarmConfig,
		ssi.backUpConfig,
	}
	for _, function := range setFunc {
		if err := function(); err != nil {
			hwlog.RunLog.Error(err)
			return err
		}
	}
	return nil
}

func (ssi *SetSystemInfoTask) prepareEdgeCoreDb() error {
	edgeCoreDbMgr := config.NewDbMgr(filepath.Join(ssi.ConfigDir, constants.EdgeCore), constants.DbEdgeCorePath)
	if err := edgeCoreDbMgr.InitDB(); err != nil {
		return errors.New("prepare edgecore database failed")
	}

	hwlog.RunLog.Info("prepare edgecore database success")
	return nil
}

func (ssi *SetSystemInfoTask) initOmDbWithTables(tables ...interface{}) error {
	ssi.dbMgr = config.NewDbMgr(filepath.Join(ssi.ConfigDir, constants.EdgeOm), constants.DbEdgeOmPath)
	if err := ssi.dbMgr.InitDB(); err != nil {
		hwlog.RunLog.Errorf("init edge om database failed, error: %v", err)
		return errors.New("init edge om database failed")
	}

	for _, table := range tables {
		if err := database.CreateTableIfNotExist(table); err != nil {
			hwlog.RunLog.Errorf("create table %T failed, error: %v", table, err)
			return errors.New("create table failed")
		}
	}
	return nil
}

func (ssi *SetSystemInfoTask) setInstallConfig() error {
	sn, err := util.GetSerialNumber(ssi.ConfigPathMgr.GetInstallRootDir())
	if err != nil {
		return fmt.Errorf("get serial number failed, error: %v", err)
	}

	hwlog.RunLog.Infof("the serial number is: %s", sn)
	installConfig := &config.InstallerConfig{
		SerialNumber:  sn,
		InstallDir:    ssi.ConfigPathMgr.GetInstallRootDir(),
		LogPath:       ssi.LogPathMgr.GetLogRootDir(),
		LogBackupPath: ssi.LogPathMgr.GetLogBackupRootDir(),
	}
	if err = config.SetInstall(ssi.dbMgr, installConfig); err != nil {
		return fmt.Errorf("set install config failed, error: %v", err)
	}

	// sn config file only be used one time in installation progress.
	if err = fileutils.DeleteFile(ssi.ConfigPathMgr.GetSnPath()); err != nil {
		hwlog.RunLog.Warnf("clean [serial-number.json] failed, please delete manually, %v", err)
	}
	hwlog.RunLog.Info("set install config success")
	return nil
}

func (ssi *SetSystemInfoTask) setEdgeCoreConfig() error {
	installConfig, err := config.GetInstall(ssi.dbMgr)
	if err != nil {
		return fmt.Errorf("get install config failed, error: %v", err)
	}

	edgeCoreConfigPath := filepath.Join(ssi.ConfigDir, constants.EdgeCore, constants.EdgeCoreJsonName)
	dataSource := ssi.ConfigPathMgr.GetEdgeCoreDbPath()
	serialNumber := installConfig.SerialNumber
	if err = config.SetDatabase(edgeCoreConfigPath, dataSource); err != nil {
		return fmt.Errorf("save database to edge core config file failed, error: %v", err)
	}
	if err = config.SetCertPath(edgeCoreConfigPath, ssi.ConfigPathMgr); err != nil {
		return fmt.Errorf("save certPath to edge core config file failed, error: %v", err)
	}
	if err = config.SetHostname(edgeCoreConfigPath, strings.ToLower(serialNumber)); err != nil {
		return fmt.Errorf("save hostnameOverride to edge core config file failed, error: %v", err)
	}
	if err = config.SetSerialNumber(edgeCoreConfigPath, serialNumber); err != nil {
		return fmt.Errorf("save serialNumber to edge core config file failed, error: %v", err)
	}
	if err = setCgroupDriver(edgeCoreConfigPath); err != nil {
		return fmt.Errorf("save cgroupDriver to edge core config file failed, error: %v", err)
	}
	hwlog.RunLog.Info("set edge core config success")
	return nil
}

func (ssi *SetSystemInfoTask) setDefaultNetConfig() error {
	netConfig := config.NetManager{
		NetType: constants.FD,
		WithOm:  true,
	}
	if err := config.SetNetManager(ssi.dbMgr, &netConfig); err != nil {
		hwlog.RunLog.Errorf("set default net config failed: %v", err)
		return err
	}
	hwlog.RunLog.Info("set default net config success")
	return nil
}

func (ssi *SetSystemInfoTask) backUpConfig() error {
	if err := backuputils.NewBackupDirMgr(ssi.ConfigDir, backuputils.CrlFileType, backuputils.CrtFileType,
		backuputils.JsonFileType, backuputils.KeyFileType).BackUp(); err != nil {
		hwlog.RunLog.Warnf("create backup files for mef config failed, error: %v", err)
		return nil
	}

	edgeMainDir := filepath.Join(ssi.ConfigDir, constants.EdgeMain)
	if err := util.SetPathOwnerGroupToMEFEdge(edgeMainDir, true, false); err != nil {
		hwlog.RunLog.Errorf("set edge-main config dir owner for backup files failed, error: %v", err)
		return err
	}
	hwlog.RunLog.Info("create backup files for mef config success")
	return nil
}

func setCgroupDriver(edgeCoreConfigPath string) error {
	const driverCgroupfs = "cgroupfs"
	cgroupDriver, err := util.GetCgroupDriver()
	if err != nil {
		hwlog.RunLog.Warn("get docker cgroup driver failed, use default value cgroupfs")
		cgroupDriver = driverCgroupfs
	}
	return config.SetCgroupDriver(edgeCoreConfigPath, cgroupDriver)
}
