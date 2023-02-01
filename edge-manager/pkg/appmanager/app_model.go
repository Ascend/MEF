// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager repository
package appmanager

import (
	"errors"
	"sync"

	"gorm.io/gorm"

	"edge-manager/pkg/database"
	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
)

const noneDealRecode = 0

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
	listAppsInfo(page, pageSize uint64, name string) ([]AppInfo, error)
	countListAppsInfo(string) (int64, error)
	countDeployedApp() (int64, int64, error)
	getNodeGroupInfosByAppID(uint64) ([]types.NodeGroupInfo, error)
	getAppInfoById(appId uint64) (*AppInfo, error)
	getAppInfoByName(string) (*AppInfo, error)
	deleteAppById(uint64) (int64, error)
	queryNodeGroup(uint64) ([]types.NodeGroupInfo, error)
	listAppInstancesById(uint64) ([]AppInstance, error)
	listAppInstancesByNode(uint64) ([]AppInstance, error)
	listAppInstances(page, pageSize uint64, name string) ([]AppInstance, error)
	countListAppInstances(string) (int64, error)
	deleteAllRemainingInstance() error
	addPod(*AppInstance) error
	updatePod(*AppInstance) error
	deletePod(*AppInstance) error
	deleteAllRemainingDaemonSet() error
	addDaemonSet(*AppDaemonSet) error
	updateDaemonSet(*AppDaemonSet) error
	deleteDaemonSet(string) error
	getNodeGroupName(appID uint64, nodeGroupID uint64) (string, error)
	countDeployedAppByGroupID(uint64) (int64, error)
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
	return a.db.Model(AppInfo{}).Create(appInfo).Error
}

func (a *AppRepositoryImpl) updateApp(appId uint64, column string, value interface{}) error {
	return a.db.Model(AppInfo{}).Where("id = ?", appId).Update(column, value).Error
}

func (a *AppRepositoryImpl) queryNodeGroup(appId uint64) ([]types.NodeGroupInfo, error) {
	var daemonSets []AppDaemonSet
	if err := a.db.Model(AppDaemonSet{}).Where("app_id = ?", appId).Find(&daemonSets).Error; err != nil {
		return nil, err
	}
	var nodeGroups []types.NodeGroupInfo
	for _, daemonSet := range daemonSets {
		nodeGroups = append(nodeGroups, types.NodeGroupInfo{
			NodeGroupID:   daemonSet.NodeGroupID,
			NodeGroupName: daemonSet.NodeGroupName,
		})
	}
	return nodeGroups, nil
}

func (a *AppRepositoryImpl) listAppsInfo(page, pageSize uint64, name string) ([]AppInfo, error) {
	var appsInfo []AppInfo
	if err := a.db.Model(AppInfo{}).Scopes(getAppInfoByLikeName(page, pageSize, name)).
		Find(&appsInfo).Error; err != nil {
		return nil, err
	}
	return appsInfo, nil
}

func (a *AppRepositoryImpl) countListAppsInfo(name string) (int64, error) {
	var totalAppInfo int64
	if err := a.db.Model(AppInfo{}).Where("INSTR(app_name, ?)", name).Count(&totalAppInfo).Error; err != nil {
		return 0, err
	}
	return totalAppInfo, nil
}

func (a *AppRepositoryImpl) countDeployedApp() (int64, int64, error) {
	var deployedAppNums, unDeployedAppNums, totalAppNums int64
	if err := a.db.Model(AppDaemonSet{}).Distinct("app_id").Count(&deployedAppNums).Error; err != nil {
		return 0, 0, err
	}
	if err := a.db.Model(AppInfo{}).Distinct("id").Count(&totalAppNums).Error; err != nil {
		return 0, 0, err
	}
	unDeployedAppNums = totalAppNums - deployedAppNums
	return deployedAppNums, unDeployedAppNums, nil
}

func (a *AppRepositoryImpl) deleteAppById(appId uint64) (int64, error) {
	err := a.db.Model(AppDaemonSet{}).Where("app_id = ?", appId).First(&AppDaemonSet{}).Error
	if err == nil {
		return noneDealRecode, errors.New("app is referenced, can not delete ")
	}
	if err != gorm.ErrRecordNotFound {
		return noneDealRecode, errors.New("find app instance failed when delete app")
	}
	rowsAffected := a.db.Model(AppInfo{}).Where("id = ?", appId).Delete(&AppInfo{})
	if rowsAffected.Error != nil {
		return rowsAffected.RowsAffected, errors.New("delete app info db error")
	}
	return rowsAffected.RowsAffected, nil
}

func (a *AppRepositoryImpl) getAppInfoById(appId uint64) (*AppInfo, error) {
	var appInfo *AppInfo
	if err := a.db.Model(AppInfo{}).Where("id = ?", appId).First(&appInfo).Error; err != nil {
		return nil, err
	}
	return appInfo, nil
}

func (a *AppRepositoryImpl) getNodeGroupInfosByAppID(appId uint64) ([]types.NodeGroupInfo, error) {
	var nodeGroupInfo []types.NodeGroupInfo
	if err := a.db.Model(AppDaemonSet{}).Where("app_id = ?", appId).Find(&nodeGroupInfo).Error; err != nil {
		return nil, err
	}
	return nodeGroupInfo, nil
}

func (a *AppRepositoryImpl) getAppInfoByName(appName string) (*AppInfo, error) {
	var appInfo *AppInfo
	if err := a.db.Model(AppInfo{}).Where("app_name = ?", appName).First(&appInfo).Error; err != nil {
		return nil, err
	}
	return appInfo, nil
}

func getAppInfoByLikeName(page, pageSize uint64, appName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(common.Paginate(page, pageSize)).Where("INSTR(app_name, ?)", appName)
	}
}

func (a *AppRepositoryImpl) listAppInstancesById(appId uint64) ([]AppInstance, error) {
	var deployedApps []AppInstance
	if err := a.db.Model(AppInstance{}).Where("app_id = ?", appId).Find(&deployedApps).Error; err != nil {
		return nil, err
	}
	return deployedApps, nil
}

func (a *AppRepositoryImpl) listAppInstancesByNode(nodeId uint64) ([]AppInstance, error) {
	var deployedApps []AppInstance
	if err := a.db.Model(AppInstance{}).Where("node_id = ?", nodeId).Find(&deployedApps).Error; err != nil {
		return nil, err
	}
	return deployedApps, nil
}

func (a *AppRepositoryImpl) listAppInstances(page, pageSize uint64, name string) ([]AppInstance, error) {
	var deployedApps []AppInstance
	if err := a.db.Model(AppInstance{}).Scopes(getAppInfoByLikeName(page, pageSize, name)).
		Find(&deployedApps).Error; err != nil {
		return nil, err
	}
	return deployedApps, nil
}

func (a *AppRepositoryImpl) countListAppInstances(name string) (int64, error) {
	var totalAppInstances int64
	if err := a.db.Model(AppInstance{}).Where("INSTR(app_name, ?)", name).
		Count(&totalAppInstances).Error; err != nil {
		return 0, err
	}
	return totalAppInstances, nil
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

func (a *AppRepositoryImpl) deleteAllRemainingDaemonSet() error {
	return a.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&AppDaemonSet{}).Error
}

func (a *AppRepositoryImpl) addDaemonSet(set *AppDaemonSet) error {
	return a.db.Model(AppDaemonSet{}).Create(set).Error
}

func (a *AppRepositoryImpl) updateDaemonSet(set *AppDaemonSet) error {
	var appDaemonSet AppDaemonSet
	a.db.Model(AppDaemonSet{}).Where("daemon_set_name = ?", set.DaemonSetName).First(&appDaemonSet)
	if appDaemonSet.NodeGroupName == set.NodeGroupName {
		return nil
	}
	return a.db.Model(AppDaemonSet{}).Updates(set).Error
}

func (a *AppRepositoryImpl) deleteDaemonSet(name string) error {
	return a.db.Model(AppDaemonSet{}).Where("daemon_set_name = ?", name).Delete(&AppDaemonSet{}).Error
}

func (a *AppRepositoryImpl) getNodeGroupName(appID uint64, nodeGroupID uint64) (string, error) {
	var appDaemonSet AppDaemonSet
	if err := a.db.Model(AppDaemonSet{}).Where("app_id = ? and node_group_id = ?", appID, nodeGroupID).
		First(&appDaemonSet).Error; err != nil {
		return "", err
	}
	return appDaemonSet.NodeGroupName, nil
}

func (a *AppRepositoryImpl) countDeployedAppByGroupID(nodeGroupID uint64) (int64, error) {
	var deployedAppCount int64
	if err := a.db.Model(AppDaemonSet{}).Where("node_group_id = ?", nodeGroupID).
		Count(&deployedAppCount).Error; err != nil {
		return 0, err
	}
	return deployedAppCount, nil
}
