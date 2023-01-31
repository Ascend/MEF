// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager the table configmap_infos operation
package appmanager

import (
	"errors"
	"strings"
	"sync"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/database"
	"huawei.com/mindxedge/base/common"
)

var (
	configmapRepositoryInitOnce sync.Once
	configmapRepository         ConfigmapRepository
)

// ConfigmapRepositoryImpl configmap service struct
type ConfigmapRepositoryImpl struct {
	db *gorm.DB
}

// ConfigmapRepository for configmap method to operate db
type ConfigmapRepository interface {
	createConfigmap(configmapInfo *ConfigmapInfo) error
	deleteConfigmapByID(configmapID int64) error
	updateConfigmapByName(configmapName string, configmapInfo *ConfigmapInfo) error
	queryConfigmapByID(configmapID int64) (*ConfigmapInfo, error)
	queryConfigmapByName(configmapName string) (*ConfigmapInfo, error)
	listConfigmapInfo(page, pageSize uint64, name string) ([]ConfigmapInfo, error)
	configmapInfosListCountByName(name string) (int64, error)
}

// ConfigmapRepositoryInstance returns the singleton instance of configmap service
func ConfigmapRepositoryInstance() ConfigmapRepository {
	configmapRepositoryInitOnce.Do(func() {
		configmapRepository = &ConfigmapRepositoryImpl{db: database.GetDb()}
	})
	return configmapRepository
}

func (ci *ConfigmapRepositoryImpl) createConfigmap(configmapInfo *ConfigmapInfo) error {
	if err := ci.db.Model(ConfigmapInfo{}).Create(configmapInfo).Error; err != nil {
		if strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
			return errors.New("configmap name is duplicate")
		}
		return err
	}

	hwlog.RunLog.Infof("create configmap [%s] in db success", configmapInfo.ConfigmapName)
	return nil
}

func (ci *ConfigmapRepositoryImpl) deleteConfigmapByID(configmapID int64) error {
	if err := ci.db.Model(ConfigmapInfo{}).Where("configmap_id = ?", configmapID).Delete(&ConfigmapInfo{}).
		Error; err != nil {
		return err
	}

	hwlog.RunLog.Infof("delete configmap [%d] from db success", configmapID)
	return nil
}

func (ci *ConfigmapRepositoryImpl) updateConfigmapByName(configmapName string, configmapInfo *ConfigmapInfo) error {
	if err := ci.db.Model(ConfigmapInfo{}).Where("configmap_name = ?", configmapName).
		Updates(&configmapInfo).Error; err != nil {
		return err
	}

	hwlog.RunLog.Infof("update configmap [%s] from db success", configmapName)
	return nil
}

func (ci *ConfigmapRepositoryImpl) queryConfigmapByID(configmapID int64) (*ConfigmapInfo, error) {
	var configmapInfo *ConfigmapInfo
	if err := ci.db.Model(ConfigmapInfo{}).Where("configmap_id = ?", configmapID).First(&configmapInfo).
		Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("configmap is nonexistent, can not be queried")
		}
		return nil, err
	}

	hwlog.RunLog.Infof("query configmap [%d] from db success", configmapID)
	return configmapInfo, nil
}

func (ci *ConfigmapRepositoryImpl) queryConfigmapByName(configmapName string) (*ConfigmapInfo, error) {
	var configmapInfo *ConfigmapInfo
	if err := ci.db.Model(ConfigmapInfo{}).Where("configmap_name = ?", configmapName).First(&configmapInfo).
		Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("configmap is nonexistent, can not be queried")
		}
		return nil, err
	}

	hwlog.RunLog.Infof("query configmap [%s] from db success", configmapName)
	return configmapInfo, nil
}

func (ci *ConfigmapRepositoryImpl) listConfigmapInfo(page, pageSize uint64, name string) ([]ConfigmapInfo, error) {
	var configmapsInfo []ConfigmapInfo
	if err := ci.db.Model(ConfigmapInfo{}).Scopes(getConfigmapInfoByLikeName(page, pageSize, name)).
		Find(&configmapsInfo).Error; err != nil {
		return nil, err
	}

	hwlog.RunLog.Info("list configmap info from db success")
	return configmapsInfo, nil
}

func getConfigmapInfoByLikeName(page, pageSize uint64, configmapName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(common.Paginate(page, pageSize)).Where("INSTR(configmap_name, ?)", configmapName)
	}
}

func (ci *ConfigmapRepositoryImpl) configmapInfosListCountByName(name string) (int64, error) {
	var totalConfigmapInfo int64
	if err := ci.db.Model(ConfigmapInfo{}).Where("INSTR(configmap_name, ?)", name).
		Count(&totalConfigmapInfo).Error; err != nil {
		return 0, err
	}

	return totalConfigmapInfo, nil
}
