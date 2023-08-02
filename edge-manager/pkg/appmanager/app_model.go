// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager repository
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
	updateApp(*AppInfo) error
	listAppsInfo(page, pageSize uint64, name string) ([]AppInfo, error)
	countListAppsInfo(string) (int64, error)
	countDeployedApp() (int64, int64, error)
	getNodeGroupInfosByAppID(uint64) ([]types.NodeGroupInfo, error)
	getAppInfoById(appId uint64) (*AppInfo, error)
	getAppInfoByName(string) (*AppInfo, error)
	deleteAppById(uint64) (int64, error)
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

	createAppAndUpdateCm(req *CreateAppReq) (uint64, error)
	deleteSingleApp(appId uint64) error

	isAppReferenced(appId uint64) error
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

func (a *AppRepositoryImpl) updateApp(appInfo *AppInfo) error {
	return a.db.Transaction(func(tx *gorm.DB) error {
		if stmt := tx.Model(AppInfo{}).Where("id = ?", appInfo.ID).
			Update("containers", appInfo.Containers); stmt.Error != nil {
			return errors.New("update app to db failed")
		}

		var daemonSets []AppDaemonSet
		if stmt := tx.Model(AppDaemonSet{}).Where("app_id = ?", appInfo.ID).Find(&daemonSets); stmt.Error != nil {
			return errors.New("get node group failed ")
		}
		var nodeGroups []types.NodeGroupInfo
		for _, daemonSet := range daemonSets {
			nodeGroups = append(nodeGroups, types.NodeGroupInfo{
				NodeGroupID:   daemonSet.NodeGroupID,
				NodeGroupName: daemonSet.NodeGroupName,
			})
		}

		if err := updateNodeGroupDaemonSet(appInfo, nodeGroups); err != nil {
			return fmt.Errorf("update node group daemon set failed: %s", err.Error())
		}
		return nil
	})
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
		return noneDealRecode, errors.New("app is referenced, can not be deleted")
	}
	if err != gorm.ErrRecordNotFound {
		return noneDealRecode, errors.New("find app instance failed when deleting app")
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

func (a *AppRepositoryImpl) isAppReferenced(appId uint64) error {
	err := a.db.Model(AppDaemonSet{}).Where("app_id = ?", appId).First(&AppDaemonSet{}).Error
	if err == nil {
		hwlog.RunLog.Error("app is referenced, can not be deleted")
		return errors.New("app is referenced, can not be deleted")
	}
	if err != gorm.ErrRecordNotFound {
		hwlog.RunLog.Errorf("find app instance failed when deleting app, error: %v", err)
		return errors.New("find app instance failed when deleting app")
	}
	return nil
}

func (a *AppRepositoryImpl) createAppAndUpdateCm(req *CreateAppReq) (uint64, error) {
	if req == nil {
		hwlog.RunLog.Error("create app req is nil")
		return 0, errors.New("create app req is nil")
	}

	app, err := req.toDb()
	if err != nil {
		hwlog.RunLog.Errorf("convert app request param to db failed, error: %v", err)
		return 0, errors.New("convert app request param to db failed")
	}

	return app.ID, a.db.Transaction(func(tx *gorm.DB) error {
		return createAppAndUpdateCm(tx, app, req)
	})
}

func createAppAndUpdateCm(tx *gorm.DB, app *AppInfo, req *CreateAppReq) error {
	// create app
	if err := tx.Model(AppInfo{}).Create(app).Error; err != nil {
		if strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
			hwlog.RunLog.Error("app name is duplicate")
			return errors.New(common.ErrDbUniqueFailed)
		}
		hwlog.RunLog.Errorf("create app [%s] in db failed, error: %v", app.AppName, err)
		return errors.New("create app in db failed")
	}

	for _, cmName := range getCmsInApp(req.Containers) {
		// query configmap
		var cmInfo ConfigmapInfo
		if err := tx.Model(ConfigmapInfo{}).Where("configmap_name = ?", cmName).First(&cmInfo).Error; err != nil {
			hwlog.RunLog.Errorf("query configmap name failed, error: %v", err)
			return errors.New("query configmap name failed")
		}

		// update configmap associated app list
		var appList []uint64
		if cmInfo.AssociatedAppList != "" { // 此处若直接对空切片进行反序列化：unexpected end of JSON input
			if err := json.Unmarshal([]byte(cmInfo.AssociatedAppList), &appList); err != nil {
				hwlog.RunLog.Errorf("unmarshal associated app list failed, error: %v", err)
				return errors.New("unmarshal associated app list failed")
			}
		}

		appList = append(appList, app.ID)
		appByte, err := json.Marshal(appList)
		if err != nil {
			hwlog.RunLog.Errorf("marshal associated app list failed, error: %v", err)
			return errors.New("marshal associated app list failed")
		}
		cmInfo.AssociatedAppList = string(appByte)

		stmt := tx.Model(&ConfigmapInfo{}).Where("configmap_name = ?", cmName).Updates(&cmInfo)
		if stmt.Error != nil { // 此处不可能出现“UNIQUE constraint failed”，因此无需判断
			hwlog.RunLog.Errorf("update configmap [%d] to db failed, error: %v", cmInfo.ID, stmt.Error)
			return errors.New("update configmap to db failed")
		}
		if stmt.RowsAffected != 1 {
			hwlog.RunLog.Errorf("update configmap [%d] to db failed, rows affected wrong", cmInfo.ID)
			return errors.New("update configmap to db failed")
		}
	}

	return nil
}

func (a *AppRepositoryImpl) deleteSingleApp(appId uint64) error {
	return a.db.Transaction(func(tx *gorm.DB) error {
		return deleteSingleApp(tx, appId)
	})
}

func deleteSingleApp(tx *gorm.DB, appId uint64) error {
	var appInfo AppInfo
	if err := tx.Model(AppInfo{}).Where("id = ?", appId).First(&appInfo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			hwlog.RunLog.Errorf("app [%d] does not exist", appId)
			return fmt.Errorf("app [%d] does not exist", appId)
		}
		hwlog.RunLog.Errorf("query app [%d] from db failed, error: %v", appId, err)
		return fmt.Errorf("query app [%d] from db failed", appId)
	}

	// delete app
	stmt := tx.Model(AppInfo{}).Where("id = ?", appId).Delete(&AppInfo{})
	if stmt.Error != nil {
		hwlog.RunLog.Errorf("delete app info from db error, error: %v", stmt.Error)
		return errors.New("delete app info from db error")
	}
	if stmt.RowsAffected != 1 {
		hwlog.RunLog.Errorf("app id [%d] does not exist", appId)
		return fmt.Errorf("app id [%d] does not exist", appId)
	}

	var containerList []Container
	if err := json.Unmarshal([]byte(appInfo.Containers), &containerList); err != nil {
		hwlog.RunLog.Errorf("unmarshal app containers failed, error: %v", err)
		return errors.New("unmarshal app containers failed")
	}

	for _, cmName := range getCmsInApp(containerList) {
		// query configmap and delete app from list
		if err := updateSingleCm(tx, cmName, appInfo.ID); err != nil {
			return err
		}
	}

	return nil
}

func getCmsInApp(containerList []Container) []string {
	// 一个app内不同容器使用同一个configmap挂载卷，也只算一个
	cmNamesMap := make(map[string]struct{})
	var cmNames []string

	for _, container := range containerList {
		for _, cmVolume := range container.ConfigmapVolumes {
			if _, ok := cmNamesMap[cmVolume.ConfigmapName]; !ok {
				cmNamesMap[cmVolume.ConfigmapName] = struct{}{}
				cmNames = append(cmNames, cmVolume.ConfigmapName)
				continue
			}
		}
	}

	return cmNames
}

func updateSingleCm(tx *gorm.DB, cmName string, appId uint64) error {
	// query configmap info
	var cmInfo ConfigmapInfo
	if err := tx.Model(ConfigmapInfo{}).Where("configmap_name = ?", cmName).First(&cmInfo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			hwlog.RunLog.Errorf("configmap [%s] does not exist", cmName)
			return fmt.Errorf("configmap [%s] does not exist", cmName)
		}
		hwlog.RunLog.Errorf("query configmap [%s] from db failed, error: %v", cmName, err)
		return fmt.Errorf("query configmap [%s] from db failed", cmName)
	}

	// update configmap associated app list
	updateCmInfo, err := updateAppList(&cmInfo, appId)
	if err != nil {
		return err
	}

	stmt := tx.Model(&ConfigmapInfo{}).Where("configmap_name = ?", cmName).Updates(&updateCmInfo)
	if stmt.Error != nil {
		if strings.Contains(stmt.Error.Error(), common.ErrDbUniqueFailed) {
			hwlog.RunLog.Errorf("configmap name [%s] is duplicate", cmName)
			return fmt.Errorf("configmap name [%s] is duplicate", cmName)
		}
		hwlog.RunLog.Errorf("update configmap [%s] to db failed, error: %v", cmName, stmt.Error)
		return errors.New("update configmap to db failed")
	}
	if stmt.RowsAffected != 1 {
		hwlog.RunLog.Errorf("update configmap [%s] to db failed, rows affected wrong", cmName)
		return errors.New("update configmap to db failed")
	}

	return nil
}

func updateAppList(cmInfo *ConfigmapInfo, appId uint64) (*ConfigmapInfo, error) {
	if cmInfo == nil {
		return nil, errors.New("configmap info is nil")
	}

	var appList []uint64
	if err := json.Unmarshal([]byte(cmInfo.AssociatedAppList), &appList); err != nil {
		hwlog.RunLog.Errorf("unmarshal associated app list failed, error: %v", err)
		return nil, errors.New("unmarshal associated app list failed")
	}
	appList = deleteAppIdFromList(appId, appList)

	appByte, err := json.Marshal(appList)
	if err != nil {
		hwlog.RunLog.Errorf("marshal associated app list failed, error: %v", err)
		return nil, errors.New("marshal associated app list failed")
	}
	cmInfo.AssociatedAppList = string(appByte)
	return cmInfo, nil
}

func deleteAppIdFromList(appId uint64, appList []uint64) []uint64 {
	var idIndex int
	for i := range appList {
		if appList[i] == appId {
			idIndex = i
			break
		}
	}
	appList = append(appList[:idIndex], appList[idIndex+1:]...)
	return appList
}
