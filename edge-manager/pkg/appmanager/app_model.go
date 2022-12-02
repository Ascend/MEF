// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager repository
package appmanager

import (
	"edge-manager/pkg/util"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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
	AppInfo AppInfo `json:"appInfo"`
	// NodeGroup is node group information of app
	NodeGroup nodemanager.NodeGroup `json:"nodeGroup"`
}

type CreateReturnInfo struct {
	AppId uint64 `json:"appId"`
}

// ListReturnInfo encapsulate app list
type ListReturnInfo struct {
	// AppInfo is app information
	AppInfo []AppReturnInfo `json:"appInfo"`
	// Total is num of appInfos
	Total int64 `json:"total"`
}

// AppReturnInfo encapsulate app information for return
type AppReturnInfo struct {
	AppId         uint64           `json:"appId"`
	AppName       string           `json:"appName"`
	Description   string           `json:"description"`
	CreatedAt     string           `json:"createdAt"`
	ModifiedAt    string           `json:"modifiedAt"`
	NodeGroupName string           `json:"nodeGroupName"`
	NodeGroupId   []int64          `json:"nodeGroupId"`
	Containers    []util.Container `json:"containers"`
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
	listAppsInfo(uint64, uint64, string) (*ListReturnInfo, error)
	countListAppsInfo(string) (int64, error)
	getGroupNameByAppName(string) (string, error)
	getAppInfoById(appId uint64) (*AppInfo, error)
	getAppInstanceByIdAndGroup(uint64, string) (*AppInstance, error)
	getAppInfoByName(string) (*AppInfo, error)
	deployApp(*AppInstance) error
	deleteApp(uint64) error
	deleteAppInstanceByIdAndGroup(uint64, string) error
	queryNodeGroup(uint64) ([]util.NodeGroupInfo, error)
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
		hwlog.RunLog.Errorf("create app failed: %s", err.Error())
		return err
	}
	return nil
}

func (a *AppRepositoryImpl) updateApp(appId uint64, column string, value interface{}) error {
	if err := a.db.Model(AppInfo{}).Where("id = ?", appId).Update(column, value).Error; err != nil {
		hwlog.RunLog.Errorf("update app failed: %s", err.Error())
		return err
	}
	return nil
}

func (a *AppRepositoryImpl) queryApp(appId uint64) (AppInfo, error) {
	var appInfo *AppInfo
	var err error
	if appInfo, err = a.getAppInfoById(appId); err != nil {
		hwlog.RunLog.Errorf("query app id [%d] info from db failed", appId)
		return *appInfo, fmt.Errorf("query app id [%d] info from db failed", appId)
	}

	return *appInfo, nil
}

func (a *AppRepositoryImpl) queryNodeGroup(appId uint64) ([]util.NodeGroupInfo, error) {
	var appInstances []AppInstance
	if err := a.db.Model(AppInstance{}).Where("app_id = ?", appId).Find(&appInstances).Error; err != nil {
		hwlog.RunLog.Error("get appInstance db failed when query")
		return nil, err
	}
	var nodeGroups []util.NodeGroupInfo
	for _, appInstance := range appInstances {
		nodeGroups = append(nodeGroups, util.NodeGroupInfo{
			NodeGroupID:   appInstance.NodeGroupID,
			NodeGroupName: appInstance.NodeGroupName,
		})
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
		var containers []util.Container
		if err := json.Unmarshal([]byte(app.Containers), &containers); err != nil {
			hwlog.RunLog.Error("containers unmarshal failed")
			return nil, err
		}
		appInstances, err := a.queryNodeGroup(app.ID)
		if err != nil {
			hwlog.RunLog.Error("get node group name failed when list")
			return nil, err
		}
		nodeGroupName := ""
		var nodeGroupIDs []int64
		for _, appInstance := range appInstances {
			nodeGroupName += appInstance.NodeGroupName + ";"
			nodeGroupIDs = append(nodeGroupIDs, appInstance.NodeGroupID)
		}
		appReturnInfos = append(appReturnInfos, AppReturnInfo{
			AppId:         app.ID,
			AppName:       app.AppName,
			Description:   app.Description,
			NodeGroupName: strings.TrimRight(nodeGroupName, ";"),
			NodeGroupId:   nodeGroupIDs,
			CreatedAt:     app.CreatedAt,
			ModifiedAt:    app.ModifiedAt,
			Containers:    containers,
		})
	}
	return &ListReturnInfo{
		AppInfo: appReturnInfos,
	}, nil
}

// listAppsInfo return appInfo list from SQL
func (a *AppRepositoryImpl) countListAppsInfo(name string) (int64, error) {
	var totalAppInfo int64
	if err := a.db.Model(AppInfo{}).Where("app_name like ?", "%"+name+"%").Count(&totalAppInfo).Error; err != nil {
		hwlog.RunLog.Error("count list appInfo db failed")
		return 0, err
	}
	return totalAppInfo, nil
}

// deployApp deploy app on node group
func (a *AppRepositoryImpl) deployApp(appInstance *AppInstance) error {
	return a.db.Model(AppInstance{}).Create(appInstance).Error
}

// deleteApp delete app by name
func (a *AppRepositoryImpl) deleteApp(appId uint64) error {
	appInfo, err := a.getAppInfoById(appId)
	if err != nil {
		hwlog.RunLog.Error("get appInfo failed when delete")
		return err
	}
	err = a.db.Where("app_name = ?", appInfo.AppName).First(&AppInstance{}).Error
	if err != gorm.ErrRecordNotFound {
		if err == nil {
			hwlog.RunLog.Error("app is referenced, can not delete")
			return errors.New("app is referenced, can not delete ")
		}
		if err != nil {
			hwlog.RunLog.Error("find app instance failed when delete app")
			return errors.New("find app instance failed when delete app")
		}
	}

	return a.db.Model(AppInfo{}).Delete(appInfo).Error
}

func (a *AppRepositoryImpl) getAppInfoById(appId uint64) (*AppInfo, error) {
	var appInfo *AppInfo
	if err := a.db.Model(AppInfo{}).Where("id = ?", appId).First(&appInfo).Error; err != nil {
		hwlog.RunLog.Error("find app from db by id failed")
		return nil, err
	}
	return appInfo, nil
}

func (a *AppRepositoryImpl) getGroupNameByAppName(appName string) (string, error) {
	var appInstance AppInstance
	if err := a.db.Model(AppInfo{}).Where("app_name = ?", appName).First(&appInstance).Error; err != nil {
		hwlog.RunLog.Error("find app instance from db when delete app")
		return "", err
	}
	return appInstance.NodeGroupName, nil
}

func (a *AppRepositoryImpl) getAppInfoByName(appName string) (*AppInfo, error) {
	var appInfo *AppInfo
	if err := a.db.Model(AppInfo{}).Where("app_name = ?", appName).First(&appInfo).Error; err != nil {
		hwlog.RunLog.Error("find app db by name failed")
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

func (a *AppRepositoryImpl) getAppInstanceByIdAndGroup(appId uint64, nodeGroupName string) (*AppInstance, error) {
	var appInstance *AppInstance
	if err := a.db.Model(AppInstance{}).Where("app_id = ? and node_group_name = ?", appId, nodeGroupName).
		First(&appInstance).Error; err != nil {
		hwlog.RunLog.Error("find app instance from db by id failed")
		return nil, err
	}
	return appInstance, nil
}

func (a *AppRepositoryImpl) deleteAppInstanceByIdAndGroup(appId uint64, nodeGroupName string) error {
	if err := a.db.Model(AppInstance{}).Where("app_id = ? and node_group_name = ?", appId, nodeGroupName).
		Delete(AppInstance{}).Error; err != nil {
		hwlog.RunLog.Error("delete app instance from db by id failed")
		return err
	}
	return nil
}
