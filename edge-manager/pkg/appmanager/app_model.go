// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager repository
package appmanager

import (
	"edge-manager/pkg/util"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/database"
	"edge-manager/pkg/nodemanager"
)

// AppInstanceInfo encapsulate app instance information
type AppInstanceInfo struct {
	// AppInfo is app information
	AppInfo AppInfo
	// NodeGroup is node group information of app
	NodeGroup nodemanager.NodeGroup
}

// ListReturnInfo encapsulate app list
type ListReturnInfo struct {
	// AppInfo is app information
	AppInfo []AppReturnInfo
	// Total is num of appInfos
	Total int
}

// AppReturnInfo encapsulate app information for return
type AppReturnInfo struct {
	ID            uint64              `json:"id"`
	AppName       string              `json:"appName"`
	Version       string              `json:"version"`
	Description   string              `json:"description"`
	CreatedAt     string              `json:"createdAt"`
	ModifiedAt    string              `json:"modifiedAt"`
	NodeGroupName string              `json:"nodeGroupName"`
	Containers    []util.ContainerReq `json:"containers"`
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
	createApp(*AppInfo) error
	queryApp(appId string) (AppInfo, error)
	listAppsInfo(uint64, uint64, string) (*ListReturnInfo, error)
	getAppAndNodeGroupInfo(string, string) (*AppInstanceInfo, error)
	deployApp(*AppInstance) error
	deleteApp(string) error
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
func (a *AppRepositoryImpl) createApp(appInfo *AppInfo) error {
	if err := a.db.Model(AppInfo{}).Create(appInfo).Error; err != nil {
		hwlog.RunLog.Error("create appInfo db failed")
		return err
	}
	return nil
}

// queryApp query application from db
func (a *AppRepositoryImpl) queryApp(appId string) (AppInfo, error) {
	var appInfo *AppInfo
	var err error
	if appInfo, err = a.getAppInfoById(appId); err != nil {
		hwlog.RunLog.Errorf("query app id [%s] info from db failed", appId)
		return *appInfo, fmt.Errorf("query app id [%s] info from db failed", appId)
	}

	return *appInfo, nil
}

// listAppsInfo return appInfo list from SQL
func (a *AppRepositoryImpl) listAppsInfo(page, pageSize uint64, name string) (*ListReturnInfo, error) {
	var appsInfo []AppInfo
	if err := a.db.Model(AppInfo{}).Scopes(getAppInfoByLikeName(page, pageSize, name)).Find(&appsInfo).Error; err != nil {
		hwlog.RunLog.Error("list appInfo db failed")
		return nil, err
	}
	var appReturnInfos []AppReturnInfo
	for _, app := range appsInfo {
		var containers []util.ContainerReq
		if err := json.Unmarshal([]byte(app.Containers), &containers); err != nil {
			hwlog.RunLog.Error("containers unmarshal failed")
			return nil, err
		}
		appReturnInfos = append(appReturnInfos, AppReturnInfo{
			ID:            app.ID,
			AppName:       app.AppName,
			Version:       app.Version,
			Description:   app.Description,
			CreatedAt:     app.CreatedAt,
			ModifiedAt:    app.ModifiedAt,
			NodeGroupName: app.NodeGroupName,
			Containers:    containers,
		})
	}
	return &ListReturnInfo{
		AppInfo: appReturnInfos,
		Total:   len(appsInfo),
	}, nil
}

// getAppAndNodeGroupInfo get application and node group information for deploy
func (a *AppRepositoryImpl) getAppAndNodeGroupInfo(appName string, nodeGroupName string) (*AppInstanceInfo, error) {
	appInfo, err := a.getAppInfoByName(appName)
	if err != nil {
		hwlog.RunLog.Error("get appInfo failed when deploy")
		return nil, err
	}
	var nodeGroup nodemanager.NodeGroup
	if err := a.db.Model(nodemanager.NodeGroup{}).
		Where("group_name = ?", nodeGroupName).First(&nodeGroup).Error; err != nil {
		hwlog.RunLog.Error("find nodeGroup db failed")
		return nil, err
	}
	return &AppInstanceInfo{
		AppInfo:   *appInfo,
		NodeGroup: nodeGroup,
	}, nil
}

// deployApp deploy app on node group
func (a *AppRepositoryImpl) deployApp(appInstance *AppInstance) error {
	return a.db.Model(AppInstance{}).Create(appInstance).Error
}

// deleteApp delete app by name
func (a *AppRepositoryImpl) deleteApp(appName string) error {
	appInfo, err := a.getAppInfoByName(appName)
	if err != nil {
		hwlog.RunLog.Error("get appInfo failed when delete")
		return err
	}
	if appInfo.NodeGroupName != "" {
		hwlog.RunLog.Error("app is referenced, can not delete")
		return errors.New("app is referenced, can not delete ")
	}
	return a.db.Model(AppInfo{}).Delete(appInfo).Error
}

func (a *AppRepositoryImpl) getAppInfoById(appId string) (*AppInfo, error) {
	var appInfo *AppInfo
	if err := a.db.Model(AppInfo{}).Where("id = ?", appId).First(&appInfo).Error; err != nil {
		hwlog.RunLog.Error("find app from db by id failed")
		return nil, err
	}
	return appInfo, nil
}

func (a *AppRepositoryImpl) getAppInfoByName(appName string) (*AppInfo, error) {
	var appInfo *AppInfo
	if err := a.db.Model(AppInfo{}).Where("app_name = ?", appName).First(&appInfo).Error; err != nil {
		hwlog.RunLog.Error("find app db failed")
		return nil, err
	}
	return appInfo, nil
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
		return db.Scopes(paginate(page, pageSize)).Where("app_name like ?", "%"+appName+"%")
	}
}
