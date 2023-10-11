// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager the table configmap_infos operation
package appmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"gorm.io/gorm"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/kubeclient"

	"huawei.com/mindxedge/base/common"
)

var (
	configmapRepositoryInitOnce sync.Once
	configmapRepository         ConfigmapRepository
)

// CmRepositoryImpl configmap service struct
type CmRepositoryImpl struct {
}

// ConfigmapRepository for configmap method to operate db
type ConfigmapRepository interface {
	createCmInDB(configmapInfo *ConfigmapInfo) error
	deleteCmByID(configmapID uint64) (int64, error)
	updateCmByName(configmapName string, configmapInfo *ConfigmapInfo) (int64, error)
	queryCmByID(configmapID uint64) (*ConfigmapInfo, error)
	queryCmByName(configmapName string) (*ConfigmapInfo, error)
	listCmInfo(page, pageSize uint64, name string) ([]ConfigmapInfo, error)
	cmListCountByName(name string) (int64, error)

	createCm(createCmReq *ConfigmapReq) (uint64, error)
	updateCm(updateCmReq *ConfigmapReq) error
	deleteSingleCm(configmapID uint64) error
}

// CmRepositoryInstance returns the singleton instance of configmap service
func CmRepositoryInstance() ConfigmapRepository {
	configmapRepositoryInitOnce.Do(func() {
		configmapRepository = &CmRepositoryImpl{}
	})
	return configmapRepository
}

func (ci *CmRepositoryImpl) db() *gorm.DB {
	return database.GetDb()
}

func (ci *CmRepositoryImpl) createCmInDB(configmapInfo *ConfigmapInfo) error {
	return ci.db().Model(ConfigmapInfo{}).Create(configmapInfo).Error
}

func (ci *CmRepositoryImpl) deleteCmByID(configmapID uint64) (int64, error) {
	stmt := ci.db().Model(ConfigmapInfo{}).Where("id = ?", configmapID).Delete(&ConfigmapInfo{})
	return stmt.RowsAffected, stmt.Error
}

func (ci *CmRepositoryImpl) updateCmByName(configmapName string, configmapInfo *ConfigmapInfo) (int64, error) {
	stmt := ci.db().Model(&ConfigmapInfo{}).Where("configmap_name = ?", configmapName).Updates(&configmapInfo)
	return stmt.RowsAffected, stmt.Error
}

func (ci *CmRepositoryImpl) queryCmByID(configmapID uint64) (*ConfigmapInfo, error) {
	var configmapInfo *ConfigmapInfo
	return configmapInfo,
		ci.db().Model(ConfigmapInfo{}).Where("id = ?", configmapID).First(&configmapInfo).Error
}

func (ci *CmRepositoryImpl) queryCmByName(configmapName string) (*ConfigmapInfo, error) {
	var configmapInfo *ConfigmapInfo
	return configmapInfo,
		ci.db().Model(ConfigmapInfo{}).Where("configmap_name = ?", configmapName).First(&configmapInfo).Error
}

func (ci *CmRepositoryImpl) listCmInfo(page, pageSize uint64, name string) ([]ConfigmapInfo, error) {
	var configmapsInfo []ConfigmapInfo
	return configmapsInfo,
		ci.db().Model(ConfigmapInfo{}).Scopes(getConfigmapInfoByLikeName(page, pageSize, name)).Find(&configmapsInfo).Error
}

func getConfigmapInfoByLikeName(page, pageSize uint64, configmapName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(common.Paginate(page, pageSize)).Where("INSTR(configmap_name, ?)", configmapName)
	}
}

func (ci *CmRepositoryImpl) cmListCountByName(name string) (int64, error) {
	var totalConfigmapInfo int64
	return totalConfigmapInfo,
		ci.db().Model(ConfigmapInfo{}).Where("INSTR(configmap_name, ?)", name).Count(&totalConfigmapInfo).Error
}

func (ci *CmRepositoryImpl) createCm(createCmReq *ConfigmapReq) (uint64, error) {
	if createCmReq == nil {
		hwlog.RunLog.Error("create configmap req is nil")
		return 0, errors.New("create configmap req is nil")
	}

	cm, err := createCmReq.toDb()
	if err != nil {
		hwlog.RunLog.Errorf("convert cm request param to db failed, error: %v", err)
		return 0, errors.New("convert cm request param to db failed")
	}

	return cm.ID, database.Transaction(ci.db(), func(tx *gorm.DB) error {
		// create in db
		if err = tx.Model(ConfigmapInfo{}).Create(cm).Error; err != nil {

			if strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
				hwlog.RunLog.Error("configmap name is duplicate")
				return errors.New(common.ErrDbUniqueFailed)
			}

			hwlog.RunLog.Errorf("create configmap [%s] in db failed, error: %v", createCmReq.ConfigmapName, err)
			return errors.New("create configmap in db failed")
		}

		// create to k8s
		if err = createCmToK8S(createCmReq); err != nil {
			return err
		}
		return nil
	})
}

func createCmToK8S(createCmReq *ConfigmapReq) error {
	configmapToK8S := convertCmToK8S(createCmReq)

	_, err := kubeclient.GetKubeClient().CreateConfigMap(configmapToK8S)
	if err != nil {
		hwlog.RunLog.Errorf("create configmap [%s] to k8s failed, error: %v", createCmReq.ConfigmapName, err)
		return errors.New("create configmap to k8s failed")
	}

	hwlog.RunLog.Infof("create configmap [%s] to k8s success", createCmReq.ConfigmapName)
	return nil
}

func (ci *CmRepositoryImpl) updateCm(updateCmReq *ConfigmapReq) error {
	if updateCmReq == nil {
		hwlog.RunLog.Error("update configmap req is nil")
		return errors.New("update configmap req is nil")
	}

	return database.Transaction(ci.db(), func(tx *gorm.DB) error {

		// query cm info by name
		var configmapInfo ConfigmapInfo
		err := tx.Model(ConfigmapInfo{}).Where("configmap_name = ?", updateCmReq.ConfigmapName).First(&configmapInfo).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				hwlog.RunLog.Error("configmap does not exist")
				return gorm.ErrRecordNotFound
			}
			hwlog.RunLog.Errorf("query configmap [%s] from db failed, error: %v", updateCmReq.ConfigmapName, err)
			return errors.New("query configmap from db failed")
		}

		if err = updateCmParam(&configmapInfo, updateCmReq); err != nil {
			hwlog.RunLog.Errorf("convert cm request param to db failed, error: %v", err)
			return errors.New("convert cm request param to db failed")
		}

		// update to db
		stmt := tx.Model(&ConfigmapInfo{}).Where("configmap_name = ?", updateCmReq.ConfigmapName).Updates(&configmapInfo)
		if stmt.Error != nil || stmt.RowsAffected != 1 {
			if stmt.Error != nil && strings.Contains(stmt.Error.Error(), common.ErrDbUniqueFailed) {
				hwlog.RunLog.Error("configmap name is duplicate")
				return errors.New(common.ErrDbUniqueFailed)
			}
			hwlog.RunLog.Errorf("update configmap [%d] to db failed, error: %v", configmapInfo.ID, stmt.Error)
			return errors.New("update configmap to db failed")
		}

		// update to k8s
		if err = updateCmToK8S(updateCmReq); err != nil {
			return err
		}

		return nil
	})
}

func updateCmParam(configmapInfo *ConfigmapInfo, updateCmReq *ConfigmapReq) error {
	// 只支持修改description和content
	if configmapInfo == nil {
		return errors.New("configmap info is nil")
	}
	configmapInfo.Description = updateCmReq.Description
	content, err := json.Marshal(updateCmReq.ConfigmapContent)
	if err != nil {
		return err
	}
	configmapInfo.ConfigmapContent = string(content)
	return nil
}

func updateCmToK8S(updateCmReq *ConfigmapReq) error {
	configmapK8S := convertCmToK8S(updateCmReq)

	_, err := kubeclient.GetKubeClient().UpdateConfigMap(configmapK8S)
	if err != nil {
		hwlog.RunLog.Errorf("update configmap [%s] to k8s failed, error: %v", updateCmReq.ConfigmapName, err)
		return errors.New("update configmap to k8s failed")
	}

	hwlog.RunLog.Infof("update configmap [%s] to k8s success", updateCmReq.ConfigmapName)
	return nil
}

func (ci *CmRepositoryImpl) deleteSingleCm(cmID uint64) error {
	return database.Transaction(ci.db(), func(tx *gorm.DB) error {

		// query cm by id
		var cmInfoFromDB ConfigmapInfo
		if err := tx.Model(ConfigmapInfo{}).Where("id = ?", cmID).First(&cmInfoFromDB).Error; err != nil {

			if err == gorm.ErrRecordNotFound {
				hwlog.RunLog.Errorf("configmap [%d] does not exist", cmID)
				return errors.New("configmap does not exist")
			}

			hwlog.RunLog.Errorf("query configmap [%d] from db failed, error: %v", cmID, err)
			return errors.New("query configmap from db failed")
		}

		// 初始化时为""，经过app关联后为[]
		if cmInfoFromDB.AssociatedAppList != "" && cmInfoFromDB.AssociatedAppList != "[]" {
			hwlog.RunLog.Errorf("configmap [%d] is associated with app, can not be deleted", cmInfoFromDB.ID)
			return fmt.Errorf("configmap [%d] is associated with app, can not be deleted", cmInfoFromDB.ID)
		}

		// delete cm by id
		stmt := tx.Model(ConfigmapInfo{}).Where("id = ?", cmID).Delete(&ConfigmapInfo{})
		if stmt.Error != nil || stmt.RowsAffected != 1 {
			hwlog.RunLog.Errorf("delete configmap from db failed, error: %v", stmt.Error)
			return errors.New("delete configmap from db failed")
		}

		if err := deleteCmToK8S(cmInfoFromDB.ConfigmapName, cmID); err != nil {
			return err
		}
		return nil
	})
}

func deleteCmToK8S(configmapName string, configmapID uint64) error {
	if err := kubeclient.GetKubeClient().DeleteConfigMap(configmapName); err != nil {
		hwlog.RunLog.Errorf("delete configmap [%d] to k8s failed, error: %v", configmapID, err)
		return errors.New("delete configmap to k8s failed")
	}

	hwlog.RunLog.Infof("delete configmap [%d] to k8s success", configmapID)
	return nil
}
