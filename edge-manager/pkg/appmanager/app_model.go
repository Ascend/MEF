// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager repository
package appmanager

import (
	"edge-manager/pkg/nodemanager"
	"sync"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/common"
	"edge-manager/pkg/database"
)

var (
	repositoryInitOnce sync.Once
	appRepository      AppRepository
)

// AppRepositoryImpl app service struct
type AppRepositoryImpl struct {
	db *gorm.DB
}

// AppRepository for app method to operate db
type AppRepository interface {
	createApp(*AppInfo, *AppContainer) error
	listAppsDeployed(uint64, uint64) (*[]AppInstance, error)
	getAppAndNodeGroupInfo(string, string) (*AppInstanceInfo, error)
	deployApp(*AppInstance) error
}

// GetTableCount get table count
func GetTableCount(tb interface{}) (int, error) {
	var total int64
	err := database.GetDb().Model(tb).Count(&total).Error
	if err != nil {
		return 0, err
	}
	return int(total), nil
}

// AppRepositoryInstance returns the singleton instance of application service
func AppRepositoryInstance() AppRepository {
	repositoryInitOnce.Do(func() {
		appRepository = &AppRepositoryImpl{db: database.GetDb()}
	})
	return appRepository
}

// createApp Create application Db
func (a *AppRepositoryImpl) createApp(appInfo *AppInfo, container *AppContainer) error {
	if err := a.db.Model(AppInfo{}).Create(appInfo).Error; err != nil {
		hwlog.RunLog.Error("create appInfo db failed")
		return err
	}
	if err := a.db.Model(AppContainer{}).Create(container).Error; err != nil {
		hwlog.RunLog.Error("create appContainer db failed")
		return err
	}
	return nil
}

// listAppsDeployed return appInstances list from SQL
func (a *AppRepositoryImpl) listAppsDeployed(page, pageSize uint64) (*[]AppInstance, error) {
	var appsInstances []AppInstance
	return &appsInstances, a.db.Model(AppInstance{}).Scopes(paginate(page, pageSize)).Find(&appsInstances).Error
}

// getAppAndNodeGroupInfo get application and node group information for deploy
func (a *AppRepositoryImpl) getAppAndNodeGroupInfo(appName string, nodeGroupName string) (*AppInstanceInfo, error) {
	var appInstanceInfo *AppInstanceInfo
	if err := a.db.Model(AppInfo{}).Where("app_name = ?", appName).First(&appInstanceInfo.AppInfo).Error; err != nil {
		hwlog.RunLog.Error("find appInfo db failed")
		return nil, err
	}
	if err := a.db.Model(AppContainer{}).Where("app_name = ?", appName).First(&appInstanceInfo.AppContainer).Error; err != nil {
		hwlog.RunLog.Error("find appContainer db failed")
		return nil, err
	}
	if err := a.db.Model(nodemanager.NodeGroup{}).Where("nodegroup_name = ?", nodeGroupName).First(appInstanceInfo.NodeGroup).Error; err != nil {
		hwlog.RunLog.Error("find nodeGroup db failed")
		return nil, err
	}
	return appInstanceInfo, nil
}

// deployApp deploy app on node group
func (a *AppRepositoryImpl) deployApp(appInstance *AppInstance) error {
	return a.db.Model(AppInstance{}).Create(appInstance).Error
}

func paginate(page, pageSize uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = common.DefaultPage
		}
		if pageSize > common.DefaultMaxPageSize {
			pageSize = common.DefaultMaxPageSize
		}
		offset := (page - 1) * pageSize
		return db.Offset(int(offset)).Limit(int(pageSize))
	}
}
