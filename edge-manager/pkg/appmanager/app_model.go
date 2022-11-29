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
	AppId         uint64              `json:"appId"`
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
	updateApp(appId uint64, column string, value interface{}) error
	queryApp(appId uint64) (AppInfo, error)
	queryNodeGroup(appId uint64) ([]nodemanager.NodeGroup, error)
	listAppsInfo(uint64, uint64, string) (*ListReturnInfo, error)
	getAppInfo(appId uint64) (AppInfo, error)
	getNodeGroupInfo(nodeGroupName string) (*nodemanager.NodeGroup, error)
	deployApp(*AppInstance) error
	deleteApp(appId uint64) error
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
		hwlog.RunLog.Error("create appInfo db failed, %s", err.Error())
		return err
	}
	return nil
}

// updateApp update application
func (a *AppRepositoryImpl) updateApp(appId uint64, column string, value interface{}) error {
	if err := a.db.Model(AppInfo{}).Where("id = ?", appId).Update(column, value).Error; err != nil {
		hwlog.RunLog.Errorf("update appInfo to db failed, %s", err.Error())
		return err
	}
	return nil
}

// queryApp query application from db
func (a *AppRepositoryImpl) queryApp(appId uint64) (AppInfo, error) {
	var appInfo AppInfo
	var err error
	if appInfo, err = a.getAppInfo(appId); err != nil {
		if err == gorm.ErrRecordNotFound {
			return appInfo, fmt.Errorf("app %s not exist", appId)
		}
		return appInfo, fmt.Errorf("query app id [%s] info from db failed", appId)
	}

	return appInfo, nil
}

// queryApp query node group from db
func (a *AppRepositoryImpl) queryNodeGroup(appId uint64) ([]nodemanager.NodeGroup, error) {
	var nodeGroups []nodemanager.NodeGroup
	var appInstances []AppInstance
	if err := database.GetDb().Model(AppInstance{}).Where("id = ?", appId).Find(&appInstances).Error; err != nil {
		return nil, errors.New("get app instance failed")
	}

	for _, appInstance := range appInstances {
		var nodeGroup *nodemanager.NodeGroup
		var err error
		if nodeGroup, err = a.getNodeGroupInfo(appInstance.NodeGroupName); err != nil {
			return nil, errors.New("get node group info failed")
		}
		nodeGroups = append(nodeGroups, *nodeGroup)
	}

	return nodeGroups, nil
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
			AppId:       app.ID,
			AppName:     app.AppName,
			Version:     app.Version,
			Description: app.Description,
			CreatedAt:   app.CreatedAt,
			ModifiedAt:  app.ModifiedAt,
			Containers:  containers,
		})
	}
	return &ListReturnInfo{
		AppInfo: appReturnInfos,
		Total:   len(appsInfo),
	}, nil
}

// getAppInfo get application info
func (a *AppRepositoryImpl) getAppInfo(appId uint64) (AppInfo, error) {
	var appInfo AppInfo
	if err := a.db.Model(AppInfo{}).Where("id = ?", appId).First(&appInfo).Error; err != nil {
		hwlog.RunLog.Error("find app db failed")
		return appInfo, err
	}
	return appInfo, nil
}

// getNodeGroupInfo get a group information
func (a *AppRepositoryImpl) getNodeGroupInfo(nodeGroupName string) (*nodemanager.NodeGroup, error) {
	var nodeGroup nodemanager.NodeGroup
	if err := a.db.Model(nodemanager.NodeGroup{}).
		Where("group_name = ?", nodeGroupName).First(&nodeGroup).Error; err != nil {
		hwlog.RunLog.Error("find nodeGroup failed")
		return nil, err
	}
	return &nodeGroup, nil
}

// deployApp deploy app on node group
func (a *AppRepositoryImpl) deployApp(appInstance *AppInstance) error {
	return a.db.Model(AppInstance{}).Create(appInstance).Error
}

// deleteApp delete app
func (a *AppRepositoryImpl) deleteApp(appId uint64) error {
	appInfo, err := a.getAppInfo(appId)
	if err != nil {
		hwlog.RunLog.Error("get appInfo failed when delete")
		return err
	}
	nodeGroups, err := a.queryNodeGroup(appId)
	if err != nil {
		return err
	}
	if len(nodeGroups) == 0 {
		hwlog.RunLog.Error("app is referenced, can not delete")
		return errors.New("app is referenced, can not delete ")
	}
	return a.db.Model(AppInfo{}).Delete(appInfo).Error
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
