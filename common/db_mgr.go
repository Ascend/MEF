// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package common about db manager
package common

import (
	"errors"
	"fmt"
	"path/filepath"

	"gorm.io/gorm"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
)

// DbMgr a database manager for query and setting
type DbMgr struct {
	dbDir  string
	dbName string
}

// NewDbMgr create a new database manager
func NewDbMgr(dbDir, dbName string) *DbMgr {
	return &DbMgr{
		dbDir:  dbDir,
		dbName: dbName,
	}
}

func (d *DbMgr) checkAndInitDB() error {
	if database.GetDb() != nil {
		return nil
	}
	return d.InitDB()
}

// InitDB init database
func (d *DbMgr) InitDB() error {
	dbPath := filepath.Join(d.dbDir, d.dbName)
	if err := fileutils.MakeSureDir(dbPath); err != nil {
		hwlog.RunLog.Errorf("make sure db path [%s] failed, error: %v", dbPath, err)
		return errors.New("make sure db path failed")
	}
	opts := database.Options{
		EnableBackup: true,
		BackupDbPath: dbPath + BackupDbSuffix,
		TestInterval: DbTestInterval,
	}
	if err := database.InitDB(dbPath, opts); err != nil {
		hwlog.RunLog.Errorf("init db failed, error: %v", err)
		return errors.New("init db failed")
	}
	return nil
}

// SetAlarmConfig create or update value to db
func (d *DbMgr) SetAlarmConfig(cfg *AlarmConfig) error {
	if err := d.checkAndInitDB(); err != nil {
		return err
	}

	var count int64
	if err := database.GetDb().Model(AlarmConfig{}).Where(AlarmConfig{ConfigName: cfg.ConfigName}).Count(&count).
		Error; err != nil {
		hwlog.RunLog.Error("get alarm config count failed")
		return errors.New("get alarm config count failed")
	}

	if count > 0 {
		if err := database.GetDb().Model(cfg).Updates(&cfg).Error; err != nil {
			hwlog.RunLog.Error("update alarm config failed")
			return errors.New("update alarm config failed")
		}
		return nil
	}
	if err := database.GetDb().Model(AlarmConfig{}).Create(cfg).Error; err != nil {
		hwlog.RunLog.Error("create alarm config failed")
		return errors.New("create alarm config failed")
	}
	return nil
}

// GetAlarmConfig get alarm config value from db by name
func (d *DbMgr) GetAlarmConfig(cfgName string) (int, error) {
	if err := d.checkAndInitDB(); err != nil {
		return 0, err
	}

	var alarmConfig AlarmConfig
	err := database.GetDb().Model(AlarmConfig{}).Where(AlarmConfig{ConfigName: cfgName}).First(&alarmConfig).Error
	if err == gorm.ErrRecordNotFound {
		hwlog.RunLog.Errorf("alarm config %s does not exist", cfgName)
		return 0, fmt.Errorf("alarm config %s does not exist", cfgName)
	}
	if err != nil {
		hwlog.RunLog.Error("get alarm config failed")
		return 0, errors.New("get alarm config failed")
	}
	return alarmConfig.ConfigValue, nil
}

// AlarmConfig alarm config table
type AlarmConfig struct {
	ConfigName  string `gorm:"primaryKey"`
	ConfigValue int    `gorm:"not null"`
	HasModified *bool  `gorm:"not null"`
}
