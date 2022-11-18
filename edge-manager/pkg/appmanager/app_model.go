// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager repository
package appmanager

import (
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
	listAppsByName(uint64, uint64, string) (*[]AppInfo, error)
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
		hwlog.RunLog.Infof("create appInfo db failed")
		return err
	}
	if err := a.db.Model(AppContainer{}).Create(container).Error; err != nil {
		hwlog.RunLog.Infof("create appContainer db failed")
		return err
	}
	return nil
}

// listAppsByName return applications list from SQL
func (a *AppRepositoryImpl) listAppsByName(page, pageSize uint64, appName string) (*[]AppInfo, error) {
	var apps []AppInfo
	return &apps, a.db.Scopes(getAppByLikeName(page, pageSize, appName)).Find(&apps).Error
}

func getAppByLikeName(page, pageSize uint64, appName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(paginate(page, pageSize)).Where("app_name like ?", "%"+appName+"%")
	}
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
