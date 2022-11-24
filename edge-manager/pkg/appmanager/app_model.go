// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager repository
package appmanager

import (
	"sync"

	"edge-manager/pkg/database"
	"edge-manager/pkg/nodemanager"

	"gorm.io/gorm"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

// AppInstanceInfo encapsulate app instance information
type AppInstanceInfo struct {
	// AppInfo is app information
	AppInfo AppInfo
	// AppContainer is app container information
	AppContainer AppContainer
	// NodeGroup is node group information of app
	NodeGroup nodemanager.NodeGroup
}

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
	createApp(*AppInfo, []*AppContainer) error
	listAppsInfo(uint64, uint64, string) (*[]AppInfo, error)
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
func (a *AppRepositoryImpl) createApp(appInfo *AppInfo, containers []*AppContainer) error {
	if err := a.db.Model(AppInfo{}).Create(appInfo).Error; err != nil {
		hwlog.RunLog.Error("create appInfo db failed")
		return err
	}
	for _, container := range containers {
		if err := a.db.Model(AppContainer{}).Create(container).Error; err != nil {
			hwlog.RunLog.Error("create appContainer db failed")
			return err
		}
	}
	return nil
}

// listAppsInfo return appInfo list from SQL
func (a *AppRepositoryImpl) listAppsInfo(page, pageSize uint64, name string) (*[]AppInfo, error) {
	var appsInfo []AppInfo
	return &appsInfo, a.db.Model(AppInfo{}).Scopes(getAppInfoByLikeName(page, pageSize, name)).Find(&appsInfo).Error
}

// getAppAndNodeGroupInfo get application and node group information for deploy
func (a *AppRepositoryImpl) getAppAndNodeGroupInfo(appName string, nodeGroupName string) (*AppInstanceInfo, error) {
	var appInfo AppInfo
	if err := a.db.Model(AppInfo{}).Where("app_name = ?", appName).First(&appInfo).Error; err != nil {
		hwlog.RunLog.Error("find appInfo db failed")
		return nil, err
	}
	var appContainer AppContainer
	if err := a.db.Model(AppContainer{}).
		Where("app_name = ?", appName).First(&appContainer).Error; err != nil {
		hwlog.RunLog.Error("find appContainer db failed")
		return nil, err
	}
	var nodeGroup nodemanager.NodeGroup
	if err := a.db.Model(nodemanager.NodeGroup{}).
		Where("group_name = ?", nodeGroupName).First(&nodeGroup).Error; err != nil {
		hwlog.RunLog.Error("find nodeGroup db failed")
		return nil, err
	}
	return &AppInstanceInfo{
		AppInfo:      appInfo,
		AppContainer: appContainer,
		NodeGroup:    nodeGroup,
	}, nil
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

func getAppInfoByLikeName(page, pageSize uint64, appName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(paginate(page, pageSize)).Where("node_name like ?", "%"+appName+"%")
	}
}
