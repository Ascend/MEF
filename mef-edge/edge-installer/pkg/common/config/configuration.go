// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package config to operate configuration
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"gorm.io/gorm"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
)

var (
	once sync.Once
	// NetMgr cache netManager config
	NetMgr NetManager
)

// SetNetManagerCache [method] caching net config in memory for opLog
func SetNetManagerCache(config NetManager) {
	once.Do(func() {
		NetMgr = config
	})
}

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

func (d *DbMgr) getDbPath() string {
	return filepath.Join(d.dbDir, d.dbName)
}

// InitDB init database
func (d *DbMgr) InitDB() error {
	dbPath := d.getDbPath()
	if err := fileutils.MakeSureDir(dbPath); err != nil {
		hwlog.RunLog.Errorf("create db path [%s] dir failed: %v", dbPath, err)
		return err
	}
	if err := database.InitDB(dbPath); err != nil {
		hwlog.RunLog.Errorf("init db failed: %v", err)
		return err
	}
	return nil
}

// GetConfig get value from database by name
func (d *DbMgr) GetConfig(name string, config interface{}) error {
	var configuration Configuration
	var err error
	if err = d.checkAndInitDB(); err != nil {
		return err
	}
	if err = database.GetDb().Where(Configuration{Key: name}).
		First(&configuration).Error; err == gorm.ErrRecordNotFound {
		return err
	}
	if err != nil {
		return errors.New("get configuration failed")
	}
	if err = json.Unmarshal(configuration.Value, &config); err != nil {
		return errors.New("unmarshal configuration value failed")
	}
	return nil
}

// SetConfig set value into database by name
func (d *DbMgr) SetConfig(name string, config interface{}) error {
	if err := d.checkAndInitDB(); err != nil {
		return err
	}
	value, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config failed: %v", err)
	}
	defer utils.ClearSliceByteMemory(value)
	now := time.Now().Format(constants.TimeFormat)
	configuration := Configuration{
		Key:       name,
		Value:     value,
		UpdatedAt: now,
	}
	var count int64
	if err = database.GetDb().Model(Configuration{}).Where(Configuration{Key: name}).Count(&count).
		Error; err != nil {
		return errors.New("set config failed,count error")
	}
	if count > 0 {
		if err = database.GetDb().Model(&configuration).Updates(configuration).Error; err != nil {
			return errors.New("set config failed,update error")
		}
	} else {
		configuration.CreatedAt = now
		if err = database.GetDb().Create(&configuration).Error; err != nil {
			return errors.New("set config failed,create error")
		}
	}
	return nil
}

// SetAlarmConfig create or update value into db
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
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			hwlog.RunLog.Errorf("alarm config %s does not exist", cfgName)
			return 0, fmt.Errorf("alarm config %s does not exist", cfgName)
		}
		hwlog.RunLog.Error("get alarm config failed")
		return 0, errors.New("get alarm config failed")
	}
	return alarmConfig.ConfigValue, nil
}

// GetComponentDbMgr get component db manager
func GetComponentDbMgr(component string) (*DbMgr, error) {
	compDbMap := map[string]string{
		constants.EdgeMain: constants.DbEdgeMainPath,
		constants.EdgeOm:   constants.DbEdgeOmPath,
		constants.EdgeCore: constants.DbEdgeCorePath,
	}
	dbName, ok := compDbMap[component]
	if !ok {
		hwlog.RunLog.Errorf("get component [%s] db name failed", component)
		return nil, errors.New("get component db name failed")
	}
	configPathMgr, err := path.GetConfigPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("get config path manager failed, error: %v", err)
		return nil, errors.New("get config path manager failed")
	}
	dbDir := configPathMgr.GetCompConfigDir(component)
	return NewDbMgr(dbDir, dbName), nil
}
