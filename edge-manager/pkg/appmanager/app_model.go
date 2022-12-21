// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager repository
package appmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"

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
	createApp(*AppInfo) error
	updateApp(appId uint64, column string, value interface{}) error
	queryApp(appId uint64) (AppInfo, error)
	listAppsInfo(uint64, uint64, string) (*ListReturnInfo, error)
	countListAppsInfo(string) (int64, error)
	countDeployedApp() (int64, int64, error)
	getGroupNameByAppName(string) (string, error)
	getAppInfoById(appId uint64) (*AppInfo, error)
	getAppInstanceByIdAndGroup(uint64, string) (*AppInstance, error)
	getAppInfoByName(string) (*AppInfo, error)
	deployApp(*AppInstance) error
	deleteAppById(uint64) error
	deleteAppInstanceByIdAndGroup(uint64, string) error
	queryNodeGroup(uint64) ([]NodeGroupInfo, error)
	listAppInstances(uint64) ([]AppInstance, error)
	listAppInstancesByNode(int64) ([]AppInstance, error)
	deleteAllRemainingInstance() error
	addPod(*AppInstance) error
	updatePod(*AppInstance) error
	deletePod(*AppInstance) error
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

func (a *AppRepositoryImpl) queryNodeGroup(appId uint64) ([]NodeGroupInfo, error) {
	var appInstances []AppInstance
	if err := a.db.Model(AppInstance{}).Where("app_id = ?", appId).Find(&appInstances).Error; err != nil {
		hwlog.RunLog.Error("get appInstance db failed when query")
		return nil, err
	}
	var nodeGroups []NodeGroupInfo
	for _, appInstance := range appInstances {
		nodeGroups = append(nodeGroups, NodeGroupInfo{
			NodeGroupID:   appInstance.NodeGroupID,
			NodeGroupName: appInstance.NodeGroupName,
		})
	}
	return nodeGroups, nil
}

func removeDuplicates(nodeGroupInfos []NodeGroupInfo) []NodeGroupInfo {
	result := make([]NodeGroupInfo, 0, len(nodeGroupInfos))
	temp := map[NodeGroupInfo]struct{}{}
	for _, item := range nodeGroupInfos {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func (a *AppRepositoryImpl) listAppsInfo(page, pageSize uint64, name string) (*ListReturnInfo, error) {
	var appsInfo []AppInfo
	if err := a.db.Model(AppInfo{}).Scopes(getAppInfoByLikeName(page, pageSize, name)).Find(&appsInfo).Error; err != nil {
		hwlog.RunLog.Error("list appInfo db failed")
		return nil, err
	}
	var appReturnInfos []AppReturnInfo
	for _, app := range appsInfo {
		var containers []Container
		if err := json.Unmarshal([]byte(app.Containers), &containers); err != nil {
			hwlog.RunLog.Error("containers unmarshal failed")
			return nil, err
		}
		NodeGroupInfos, err := a.queryNodeGroup(app.ID)
		if err != nil {
			hwlog.RunLog.Error("get node group name failed when list")
			return nil, err
		}

		appReturnInfos = append(appReturnInfos, AppReturnInfo{
			AppId:          app.ID,
			AppName:        app.AppName,
			Description:    app.Description,
			NodeGroupInfos: removeDuplicates(NodeGroupInfos),
			CreatedAt:      app.CreatedAt.Format(common.TimeFormat),
			ModifiedAt:     app.UpdatedAt.Format(common.TimeFormat),
			Containers:     containers,
		})
	}

	return &ListReturnInfo{
		AppInfo: appReturnInfos,
	}, nil
}

func (a *AppRepositoryImpl) countListAppsInfo(name string) (int64, error) {
	var totalAppInfo int64
	if err := a.db.Model(AppInfo{}).Where("app_name like ?", "%"+name+"%").Count(&totalAppInfo).Error; err != nil {
		hwlog.RunLog.Error("count list appInfo db failed")
		return 0, err
	}
	return totalAppInfo, nil
}

func (a *AppRepositoryImpl) countDeployedApp() (int64, int64, error) {
	var deployedAppNums, unDeployedAppNums, totalAppNums int64
	if err := a.db.Model(AppInstance{}).Distinct("app_id").Count(&deployedAppNums).Error; err != nil {
		hwlog.RunLog.Error("count deployed app db failed")
		return 0, 0, err
	}
	if err := a.db.Model(AppInfo{}).Count(&totalAppNums).Error; err != nil {
		hwlog.RunLog.Error("count all app nums db failed")
		return 0, 0, err
	}
	unDeployedAppNums = totalAppNums - deployedAppNums
	return deployedAppNums, unDeployedAppNums, nil
}

func (a *AppRepositoryImpl) deployApp(appInstance *AppInstance) error {
	return a.db.Model(AppInstance{}).Create(appInstance).Error
}

func (a *AppRepositoryImpl) deleteAppById(appId uint64) error {
	appInfo, err := a.getAppInfoById(appId)
	if err != nil {
		hwlog.RunLog.Error("get appInfo failed when delete")
		return err
	}
	err = a.db.Where("app_name = ?", appInfo.AppName).First(&AppInstance{}).Error
	if err == nil {
		hwlog.RunLog.Error("app is referenced, can not delete")
		return errors.New("app is referenced, can not delete ")
	}
	if err != gorm.ErrRecordNotFound {
		hwlog.RunLog.Error("find app instance failed when delete app")
		return errors.New("find app instance failed when delete app")
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

func (a *AppRepositoryImpl) listAppInstances(appId uint64) ([]AppInstance, error) {
	var deployedApps []AppInstance
	if err := a.db.Model(AppInstance{}).Where("app_id = ?", appId).Find(&deployedApps).Error; err != nil {
		hwlog.RunLog.Error("list app instances db failed")
		return nil, err
	}
	return deployedApps, nil
}

func (a *AppRepositoryImpl) listAppInstancesByNode(nodeId int64) ([]AppInstance, error) {
	var deployedApps []AppInstance
	if err := a.db.Model(AppInstance{}).Where("node_id = ?", nodeId).Find(&deployedApps).Error; err != nil {
		hwlog.RunLog.Error("list app instances by node db failed")
		return nil, err
	}
	return deployedApps, nil
}

func (a *AppRepositoryImpl) deleteAllRemainingInstance() error {
	return a.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&AppInstance{}).Error
}

func (a *AppRepositoryImpl) addPod(appInstance *AppInstance) error {
	return a.db.Model(AppInstance{}).Create(appInstance).Error
}

func (a *AppRepositoryImpl) updatePod(appInstance *AppInstance) error {
	var eventInstance AppInstance
	a.db.Model(AppInstance{}).Where("pod_name = ?", appInstance.PodName).First(&eventInstance)
	if eventInstance.ContainerInfo == appInstance.ContainerInfo &&
		eventInstance.NodeName == appInstance.NodeName {
		return nil
	}
	return a.db.Model(AppInstance{}).Where("pod_name = ?", appInstance.PodName).Updates(appInstance).Error
}

func (a *AppRepositoryImpl) deletePod(appInstance *AppInstance) error {
	return a.db.Model(AppInstance{}).Where("pod_name = ?", appInstance.PodName).Delete(appInstance).Error
}
