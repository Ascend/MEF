// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package tasks for effect post process
package tasks

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/common"
	"edge-installer/pkg/installer/common/tasks"
)

// PostEffectProcessTask the task for post process after upgrade
type PostEffectProcessTask struct {
	tasks.PostProcessBaseTask
	ConfigPathMgr *pathmgr.ConfigPathMgr
}

// Run post process task
func (p *PostEffectProcessTask) Run() error {
	var postFunc = []func() error{
		p.clearAlarmInDB,
		p.removeUpgradeBin,
		p.CreateSoftwareSymlink,
		p.UpdateMefServiceInfo,
		p.copyResetScriptToP7,
		p.SetSoftwareDirImmutable,
		p.smoothConfig,
		p.refreshDefaultCfgDir,
		p.backUpConfig,
		p.restart,
	}
	for _, function := range postFunc {
		if err := function(); err != nil {
			hwlog.RunLog.Error(err)
			return err
		}
	}
	return nil
}

func (p *PostEffectProcessTask) clearAlarmInDB() error {
	mgr := util.NewEdgeUGidMgr()
	defer func() {
		if err := mgr.ResetEUGid(); err != nil {
			hwlog.RunLog.Errorf("reset euid and egid failed, error: %v", err)
		}
	}()

	if err := mgr.SetEUGidToEdge(); err != nil {
		return fmt.Errorf("set euid and egid to %s failed, error: %v", constants.EdgeUserName, err)
	}

	dbPath := p.ConfigPathMgr.GetCompConfigDir(constants.EdgeMain)
	return deleteMetaByType(dbPath, constants.DbEdgeMainPath, constants.MetaAlarmKey)
}

func (p *PostEffectProcessTask) removeUpgradeBin() error {
	return p.RemoveUpgradeBinByPath(p.WorkPathMgr.GetUpgradeTempBinaryPath())
}

// backUpConfig [method] create config's and certs' backup in config dir for mef-edge running.
// This method is for smoothing old version which don't have config backup files.
func (p *PostEffectProcessTask) backUpConfig() error {
	if err := backuputils.NewBackupDirMgr(p.ConfigPathMgr.GetConfigDir(), backuputils.KeyFileType,
		backuputils.JsonFileType, backuputils.CrtFileType, backuputils.CrlFileType).BackUp(); err != nil {
		hwlog.RunLog.Warnf("create backup files for mef-config failed, error: %v", err)
	}

	edgeMainDir := p.ConfigPathMgr.GetCompConfigDir(constants.EdgeMain)
	if err := util.SetPathOwnerGroupToMEFEdge(edgeMainDir, true, false); err != nil {
		return fmt.Errorf("set [%s] confg dir owner for backup files failed, error: %v", constants.EdgeMain, err)
	}
	hwlog.RunLog.Info("create backup files for mef config success")
	return nil
}

func (p *PostEffectProcessTask) restart() error {
	if active := util.IsServiceActive(constants.EdgeOmServiceFile); !active {
		hwlog.RunLog.Info("old service is not running, no need to restart")
		return nil
	}

	hwlog.RunLog.Info("effect completed, restarting...")
	fmt.Println("effect completed, restarting...")
	if err := util.RestartService(constants.MefEdgeTargetFile); err != nil {
		hwlog.RunLog.Errorf("restart target [%s] failed, error: %v", constants.MefEdgeTargetFile, err)
		return fmt.Errorf("restart target [%s] failed", constants.MefEdgeTargetFile)
	}

	componentMgr := common.NewComponentMgr(p.WorkPathMgr.GetInstallRootDir())
	componentMgr.CheckAllServiceActive()
	hwlog.RunLog.Info("restart all services success")
	hwlog.RunLog.Info("effect edge-installer success")
	fmt.Println("effect edge-installer success")
	return nil
}

func (p *PostEffectProcessTask) smoothCommonConfig() error {
	installRootDir := p.WorkPathMgr.GetInstallRootDir()
	if err := config.SmoothEdgeCoreConfigPipePath(installRootDir, constants.NewTlsPrivateKeyFile); err != nil {
		hwlog.RunLog.Errorf("smooth pipe config to edge core config file failed, error: %v", err)
		return errors.New("smooth pipe config to edge core config file failed")
	}
	if err := config.SmoothEdgeCoreSafeConfig(installRootDir); err != nil {
		hwlog.RunLog.Errorf("smooth safe config to edge core config file failed, error: %v", err)
		return errors.New("smooth safe config to edge core config file failed")
	}
	if err := config.SmoothEdgeOmContainerConfig(installRootDir); err != nil {
		hwlog.RunLog.Errorf("smooth edge_om container config to edge om config file failed, error: %v", err)
		return errors.New("smooth edge_om container config to edge om config file failed")
	}
	return nil
}

// Meta metadata object
type Meta struct {
	Key   string `gorm:"column:key; size:256; primaryKey"`
	Type  string `gorm:"column:type; size:32"`
	Value string `gorm:"column:value; type:text"`
}

func deleteMetaByType(dbPath, dbName, typ string) error {
	dbMgr := config.NewDbMgr(dbPath, dbName)
	if err := dbMgr.InitDB(); err != nil {
		hwlog.RunLog.Errorf("init database [%s] failed: %v", dbName, err)
		return fmt.Errorf("init database [%s] failed", dbName)
	}
	if database.GetDb() == nil {
		hwlog.RunLog.Errorf("init database [%s] failed: database is nil", dbName)
		return fmt.Errorf("get database [%s] failed", dbName)
	}

	if !database.GetDb().Migrator().HasTable(&Meta{}) {
		return nil
	}
	err := database.GetDb().Model(&Meta{}).Where(&Meta{Type: typ}).First(&Meta{}).Error
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	if err != nil {
		hwlog.RunLog.Errorf("find [%s] in database failed: %v", typ, err)
		return fmt.Errorf("find [%s] in database failed", typ)
	}

	if err = database.GetDb().Model(&Meta{}).Where(&Meta{Type: typ}).Delete(&Meta{}).Error; err != nil {
		hwlog.RunLog.Errorf("delete [%s] from database failed: %v", typ, err)
		return fmt.Errorf("delete [%s] from database failed", typ)
	}
	return nil
}
