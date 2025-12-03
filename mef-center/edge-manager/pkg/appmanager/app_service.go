// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package appmanager to init app manager service
package appmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-manager/pkg/appmanager/appchecker"
	"edge-manager/pkg/config"
	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/types"
	"edge-manager/pkg/util"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/logmgmt"
)

// createApp Create application
func createApp(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start create app")

	var req CreateAppReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}

	if checkResult := appchecker.NewCreateAppChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("app create para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}
	if err := NewAppSupplementalChecker(req).Check(); err != nil {
		hwlog.RunLog.Errorf("app create para check failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: fmt.Sprintf("para check failed: %v", err)}
	}

	total, err := GetTableCount(AppInfo{})
	if err != nil {
		hwlog.RunLog.Error("get app table num failed")
		return common.RespMsg{Status: common.ErrorCheckAppMrgSize, Msg: "get app table num failed", Data: nil}
	}
	if total >= MaxApp {
		hwlog.RunLog.Error("app number is enough, can not be created")
		return common.RespMsg{Status: common.ErrorCheckAppMrgSize, Msg: "app number is enough, can not be created"}
	}

	app, err := req.toDb()
	if err != nil {
		hwlog.RunLog.Error("create app request convert to db failed")
		return common.RespMsg{Status: common.ErrorAppParamConvertDb, Msg: "get app info failed", Data: nil}
	}
	err = AppRepositoryInstance().createApp(app)
	if err != nil && strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
		hwlog.RunLog.Error("app name is duplicate")
		return common.RespMsg{Status: common.ErrorAppMrgDuplicate, Msg: "app name is duplicate", Data: nil}
	}
	if err != nil {
		hwlog.RunLog.Errorf("create app in db failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorCreateApp, Msg: "create app in db failed", Data: nil}
	}

	hwlog.RunLog.Infof("create app %s success", req.AppName)
	return common.RespMsg{Status: common.Success, Msg: "", Data: app.ID}
}

// queryApp app info
func queryApp(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start query app info")

	var appId uint64
	if err := msg.ParseContent(&appId); err != nil {
		hwlog.RunLog.Errorf("query app info failed: parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}

	if checkResult := appchecker.IdChecker().Check(appId); !checkResult.Result {
		hwlog.RunLog.Errorf("query app para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}

	appInfo, err := AppRepositoryInstance().getAppInfoById(appId)
	if err == gorm.ErrRecordNotFound {
		hwlog.RunLog.Errorf("query app id [%d] not exist", appId)
		return common.RespMsg{Status: common.ErrorAppMrgRecodeNoFound, Msg: "query app record no found", Data: nil}
	}
	if err != nil {
		hwlog.RunLog.Errorf("query app info failed %s", err.Error())
		return common.RespMsg{Status: common.ErrorQueryApp, Msg: "query app info failed", Data: nil}
	}

	nodeGroupInfoLists, err := getNodeGroupInfoList(appId)
	if err != nil {
		hwlog.RunLog.Errorf("get node group info failed %s", err.Error())
		return common.RespMsg{Status: common.ErrorQueryApp, Msg: "query app info failed", Data: nil}
	}

	resp := AppReturnInfo{
		AppID:          appInfo.ID,
		AppName:        appInfo.AppName,
		Description:    appInfo.Description,
		CreatedAt:      appInfo.CreatedAt.Format(common.TimeFormat),
		ModifiedAt:     appInfo.UpdatedAt.Format(common.TimeFormat),
		NodeGroupInfos: nodeGroupInfoLists,
	}

	if err = json.Unmarshal([]byte(appInfo.Containers), &resp.Containers); err != nil {
		hwlog.RunLog.Error("unmarshal containers info failed")
		return common.RespMsg{Status: common.ErrorUnmarshalContainer, Msg: "unmarshal containers info failed", Data: nil}
	}

	hwlog.RunLog.Info("query app success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

// listAppInfo get appInfo list
func listAppInfo(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Infof("start list app infos")

	var req types.ListReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("get apps Infos list failed: parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}

	if checkResult := util.NewPaginationQueryChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("get apps Infos list failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}

	apps, err := getListReturnInfo(req)
	if err != nil {
		hwlog.RunLog.Error("get apps Infos list failed")
		return common.RespMsg{Status: common.ErrorListApp, Msg: "get apps Infos list failed", Data: nil}
	}

	apps.Total, err = AppRepositoryInstance().countListAppsInfo(req.Name)
	if err != nil {
		hwlog.RunLog.Error("count apps Infos list failed")
		return common.RespMsg{Status: common.ErrorListApp, Msg: "count apps Infos list failed", Data: nil}
	}
	apps.Deployed, apps.UnDeployed, err = AppRepositoryInstance().countDeployedApp()
	if err != nil {
		hwlog.RunLog.Error("count deployed app failed")
		return common.RespMsg{Status: common.ErrorListApp, Msg: "count apps Infos list failed", Data: nil}
	}

	hwlog.RunLog.Info("list apps Infos success")
	return common.RespMsg{Status: common.Success, Msg: "list apps Infos success", Data: apps}
}

func getListReturnInfo(req types.ListReq) (*ListReturnInfo, error) {
	appsInfo, err := AppRepositoryInstance().listAppsInfo(req.PageNum, req.PageSize, req.Name)
	if err != nil {
		return nil, err
	}
	var appReturnInfos []AppReturnInfo
	for _, app := range appsInfo {
		var containers []Container
		if err = json.Unmarshal([]byte(app.Containers), &containers); err != nil {
			hwlog.RunLog.Error("unmarshal containers failed")
			return nil, errors.New("unmarshal containers error")
		}

		nodeGroupInfoLists, err := getNodeGroupInfoList(app.ID)
		if err != nil {
			hwlog.RunLog.Errorf("get node group list failed %s", err.Error())
			return nil, err
		}

		appReturnInfo := AppReturnInfo{
			AppID:          app.ID,
			AppName:        app.AppName,
			Description:    app.Description,
			CreatedAt:      app.CreatedAt.Format(common.TimeFormat),
			ModifiedAt:     app.UpdatedAt.Format(common.TimeFormat),
			NodeGroupInfos: nodeGroupInfoLists,
			Containers:     containers,
		}
		appReturnInfos = append(appReturnInfos, appReturnInfo)
	}
	return &ListReturnInfo{
		AppInfo: appReturnInfos,
	}, nil
}

func getNodeGroupInfoList(appId uint64) ([]types.NodeGroupInfo, error) {
	nodeGroupInfos, err := AppRepositoryInstance().getNodeGroupInfosByAppID(appId)
	if err != nil {
		return []types.NodeGroupInfo{}, err
	}
	nodeGroupIDList := make([]uint64, len(nodeGroupInfos))
	for i, nodeGroupInfo := range nodeGroupInfos {
		nodeGroupIDList[i] = nodeGroupInfo.NodeGroupID
	}
	return getNodeGroupInfos(nodeGroupIDList)
}

// deployApp deploy application on node group
func deployApp(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start deploy app")

	var req DeployAppReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}

	if checkResult := appchecker.NewDeployAppChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("deploy app para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}

	appInfo, err := AppRepositoryInstance().getAppInfoById(req.AppID)
	if err == gorm.ErrRecordNotFound {
		hwlog.RunLog.Errorf("app id [%d] not exist", req.AppID)
		return common.RespMsg{Status: common.ErrorAppMrgRecodeNoFound,
			Msg: fmt.Sprintf("app id [%d] not exist, deploy app failed", req.AppID), Data: nil}
	}
	if err != nil {
		hwlog.RunLog.Error("get app info error, deploy app failed")
		return common.RespMsg{Status: common.ErrorDeployApp, Msg: "get app info error, deploy app failed", Data: nil}
	}
	deployRes, successGroups := deployAppToNodeGroups(appInfo, req.NodeGroupIds)

	logmgmt.BatchOperationLog(fmt.Sprintf("deploy app [%s] to node groups", appInfo.AppName), successGroups)
	if len(deployRes.FailedInfos) != 0 {
		return common.RespMsg{Status: common.ErrorDeployApp, Msg: "", Data: deployRes}
	}

	hwlog.RunLog.Info("all app daemonSets create success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func deployAppToNodeGroups(appInfo *AppInfo, NodeGroupIds []uint64) (types.BatchResp, []interface{}) {
	var deployRes types.BatchResp
	var successGroups []interface{}
	failedMap := make(map[string]string)
	deployRes.FailedInfos = failedMap

	for _, nodeGroupId := range NodeGroupIds {
		groupInfo, err := getNodeGroupInfos([]uint64{nodeGroupId})
		if err != nil {
			hwlog.RunLog.Errorf("get node group [%d]'s infos failed: %v", nodeGroupId, err)
			failedMap[strconv.Itoa(int(nodeGroupId))] = fmt.Sprintf("deploy app failed, "+
				"get group info error: %v", err)
			continue
		}
		if len(groupInfo) == 0 {
			hwlog.RunLog.Errorf("get node group [%d]'s infos failed: ret is empty", nodeGroupId)
			failedMap[strconv.Itoa(int(nodeGroupId))] = "deploy app failed, get group info error: ret is empty"
			continue
		}
		groupName := groupInfo[0].NodeGroupName
		if err := deployAppToSingleNodeGroup(appInfo, nodeGroupId, groupName); err != nil {
			failedMap[strconv.Itoa(int(nodeGroupId))] = err.Error()
			continue
		}
		deployRes.SuccessIDs = append(deployRes.SuccessIDs, nodeGroupId)
		successGroups = append(successGroups, groupName)
	}
	return deployRes, successGroups
}

func deployAppToSingleNodeGroup(appInfo *AppInfo, nodeGroupId uint64, groupName string) error {
	if err := preCheckForDeployApp(appInfo.ID, nodeGroupId); err != nil {
		hwlog.RunLog.Errorf("check deploy app [%s] on node group id [%d](name=%s) failed: %s",
			appInfo.AppName, nodeGroupId, groupName, err.Error())
		return fmt.Errorf("check deploy app [%s] failed: %v", appInfo.AppName, err)
	}
	daemonSet, err := initDaemonSet(appInfo, nodeGroupId)
	if err != nil {
		hwlog.RunLog.Errorf("init daemonSet app [%s] on node group id [%d](name=%s) failed: %s",
			appInfo.AppName, nodeGroupId, groupName, err.Error())
		return fmt.Errorf("init daemonSet app [%s] failed: %v", appInfo.AppName, err)
	}
	if err := checkNodeGroupResource(nodeGroupId, daemonSet); err != nil {
		hwlog.RunLog.Errorf("check app [%s] resources on node group id [%d](name=%s) failed: %s",
			appInfo.AppName, nodeGroupId, groupName, err.Error())
		return fmt.Errorf("check app [%s] resources failed: %v", appInfo.AppName, err)
	}
	if err = appRepository.addDaemonSet(daemonSet, nodeGroupId, appInfo.ID); err != nil {
		hwlog.RunLog.Errorf("app [%s] daemonSet create on node group id [%d](name=%s) failed: %s",
			appInfo.AppName, nodeGroupId, groupName, err.Error())
		return fmt.Errorf("app [%s] daemonSet create failed: %v", appInfo.AppName, err)
	}
	if err := updateAllocatedNodeRes(daemonSet, nodeGroupId, false); err != nil {
		hwlog.RunLog.Errorf("app [%s] daemonSet create on node group id [%d](name=%s) failed, "+
			"update allocated node resource error: %s", appInfo.AppName, nodeGroupId, groupName, err.Error())
		if err := appRepository.deleteDaemonSet(daemonSet.Name); err != nil {
			hwlog.RunLog.Errorf("roll back creation for daemonSet[%s] failed", daemonSet.Name)
		}
		return fmt.Errorf("app [%s] daemonSet create failed, update allocated node resource error: %v",
			appInfo.AppName, err)
	}
	return nil
}

func preCheckForDeployApp(appId, nodeGroupId uint64) error {
	if _, err := AppRepositoryInstance().getAppDaemonSet(appId, nodeGroupId); err == nil {
		return errors.New("app already exists")
	}
	if _, err := getNodeGroupInfos([]uint64{nodeGroupId}); err != nil {
		return errors.New("group id no exist")
	}
	deployedCount, err := AppRepositoryInstance().countDeployedAppByGroupID(nodeGroupId)
	if err != nil {
		return errors.New("get deployed app count failed")
	}
	if deployedCount >= config.PodConfig.MaxDsNumberPerNodeGroup {
		return errors.New("node group out of max app limit")
	}
	return nil
}

// unDeployApp deploy application on node group
func unDeployApp(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start unDeploy app")

	var req UndeployAppReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}

	if checkResult := appchecker.NewUndeployAppChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("undeploy app para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}

	appInfo, err := AppRepositoryInstance().getAppInfoById(req.AppID)
	if err != nil {
		hwlog.RunLog.Error("get app info error, undeploy app failed")
		return common.RespMsg{Status: common.ErrorUnDeployApp, Msg: "get app info error, undeploy app failed"}
	}
	unDeployRes, successGroup := undeployAppFromNodeGroups(appInfo, req.NodeGroupIds)
	logmgmt.BatchOperationLog(fmt.Sprintf("batch delete app %s on node groups", appInfo.AppName), successGroup)

	if len(unDeployRes.FailedInfos) != 0 {
		return common.RespMsg{Status: common.ErrorUnDeployApp, Msg: "undeploy app failed", Data: unDeployRes}
	}

	hwlog.RunLog.Info("undeploy app on node group success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func undeployAppFromNodeGroups(appInfo *AppInfo, NodeGroupIds []uint64) (types.BatchResp, []interface{}) {
	var unDeployRes types.BatchResp
	var successGroup []interface{}
	failedMap := make(map[string]string)
	unDeployRes.FailedInfos = failedMap
	for _, nodeGroupId := range NodeGroupIds {
		groupInfo, err := getNodeGroupInfos([]uint64{nodeGroupId})
		if err != nil {
			hwlog.RunLog.Errorf("get node group [%d]'s infos failed: %v", nodeGroupId, err)
			failedMap[strconv.Itoa(int(nodeGroupId))] = fmt.Sprintf("get node group [%d]'s infos failed", nodeGroupId)
			continue
		}
		if len(groupInfo) == 0 {
			hwlog.RunLog.Errorf("get node group [%d]'s infos failed: ret is empty", nodeGroupId)
			failedMap[strconv.Itoa(int(nodeGroupId))] = "undeploy app failed, get group info error: ret is empty"
			continue
		}
		groupName := groupInfo[0].NodeGroupName
		if err := undeployAppFromSingleNodeGroup(appInfo, nodeGroupId, groupName); err != nil {
			failedMap[strconv.Itoa(int(nodeGroupId))] = err.Error()
			continue
		}
		unDeployRes.SuccessIDs = append(unDeployRes.SuccessIDs, nodeGroupId)
		successGroup = append(successGroup, groupInfo[0].NodeGroupName)
	}
	return unDeployRes, successGroup
}

func undeployAppFromSingleNodeGroup(appInfo *AppInfo, nodeGroupId uint64, groupName string) error {
	daemonSetName := formatDaemonSetName(appInfo.AppName, nodeGroupId)
	daemonSet, err := kubeclient.GetKubeClient().GetDaemonSet(daemonSetName)
	if err != nil {
		hwlog.RunLog.Errorf("undeploy app [%s] on node group id [%d] failed: %v", appInfo.AppName, nodeGroupId, err)
		return fmt.Errorf("undeploy app on node group id [%d] failed", nodeGroupId)
	}
	if err = appRepository.deleteDaemonSet(daemonSetName); err != nil {
		hwlog.RunLog.Errorf("undeploy app [%s] on node group id [%d] failed: %v", appInfo.AppName, nodeGroupId, err)
		return fmt.Errorf("undeploy app on node group id [%d] failed", nodeGroupId)
	}

	if err = updateAllocatedNodeRes(daemonSet, nodeGroupId, true); err != nil {
		hwlog.RunLog.Errorf("undeploy app [%s] on node group id [%d](name=%s) failed, "+
			"update allocated node resource error: %v",
			appInfo.AppName, nodeGroupId, groupName, err)
		if err = appRepository.addDaemonSet(daemonSet, nodeGroupId, appInfo.ID); err != nil {
			hwlog.RunLog.Errorf("roll back deletion for daemonSet[%s] failed", daemonSet.Name)
		}
		return fmt.Errorf("undeploy app failed, update allocated node resource error: %v", err)
	}
	return nil
}

func updateNodeGroupDaemonSet(appInfo *AppInfo, nodeGroups []types.NodeGroupInfo) error {
	for _, nodeGroup := range nodeGroups {
		daemonSet, err := initDaemonSet(appInfo, nodeGroup.NodeGroupID)
		if err != nil {
			return fmt.Errorf("init daemon set failded: %s", err.Error())
		}
		daemonSet, err = kubeclient.GetKubeClient().UpdateDaemonSet(daemonSet)
		if err != nil {
			return fmt.Errorf("update daemon set failded: %s", err.Error())
		}
	}

	return nil
}

func modifyContainerPara(req *UpdateAppReq, appInfo *AppInfo) error {
	var containers []Container
	if err := json.Unmarshal([]byte(appInfo.Containers), &containers); err != nil {
		return errors.New("unmarshal containers info failed")
	}

	if len(req.Containers) != len(containers) {
		return errors.New("container count is not equal")
	}

	for i := range req.Containers {
		containers[i].Image = req.Containers[i].Image
		containers[i].ImageVersion = req.Containers[i].ImageVersion
	}

	content, err := json.Marshal(containers)
	if err != nil {
		return errors.New("marshal containers info failed")
	}

	appInfo.Containers = string(content)

	return nil
}

// updateApp update application
func updateApp(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start update app")

	var req UpdateAppReq
	var err error
	if err = msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}

	if checkResult := appchecker.NewUpdateAppChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("app update para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}
	if err = NewAppSupplementalChecker(req.CreateAppReq).Check(); err != nil {
		hwlog.RunLog.Errorf("app create para check failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamInvalid,
			Msg: fmt.Sprintf("app create para check failed: %v", err), Data: nil}
	}

	appInfo, err := AppRepositoryInstance().getAppInfoById(req.AppID)
	if err == gorm.ErrRecordNotFound {
		hwlog.RunLog.Error("app info not exist, update failed")
		return common.RespMsg{Status: common.ErrorAppMrgRecodeNoFound, Msg: "app info not exist, update failed"}
	}
	if err != nil {
		hwlog.RunLog.Error("get app info for app update, db failed")
		return common.RespMsg{Status: common.ErrorUpdateApp, Msg: "get app info for app update, db failed", Data: nil}
	}

	if err = modifyContainerPara(&req, appInfo); err != nil {
		hwlog.RunLog.Errorf("modify app info failed: %s", err.Error())
		return common.RespMsg{Status: common.ErrorUpdateApp, Msg: "update app info failed", Data: nil}
	}
	if err = AppRepositoryInstance().updateApp(appInfo); err != nil {
		hwlog.RunLog.Errorf("update app to db failed, %v", err.Error())
		return common.RespMsg{Status: common.ErrorUpdateApp, Msg: err.Error(), Data: nil}
	}

	hwlog.RunLog.Info("app daemonSet update success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

// deleteApp delete application by appName
func deleteApp(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start delete app")

	var req DeleteAppReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}

	if checkResult := appchecker.NewDeleteAppChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("app delete para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}

	var deleteRes types.BatchResp
	var successApps []interface{}
	failedMap := make(map[string]string)
	deleteRes.FailedInfos = failedMap

	for _, appId := range req.AppIDs {
		appInfo, err := AppRepositoryInstance().getAppInfoById(appId)
		if err != nil {
			hwlog.RunLog.Errorf("get app [%d]'s info failed:%v", appId, err)
			failedMap[strconv.Itoa(int(appId))] = fmt.Sprintf("delete app failed: get app info failed")
			continue
		}
		rowsAffected, err := AppRepositoryInstance().deleteAppById(appId)
		if err != nil {
			hwlog.RunLog.Errorf("delete app [%d](name=%s) failed: %v", appId, appInfo.AppName, err)
			failedMap[strconv.Itoa(int(appId))] = fmt.Sprintf("delete app failed: %v", err)
			continue
		}
		if rowsAffected != 1 {
			hwlog.RunLog.Errorf("delete app [%d](name=%s) failed: id does not exist", appId, appInfo.AppName)
			failedMap[strconv.Itoa(int(appId))] = fmt.Sprintf("delete app failed: id does not exist")
			continue
		}
		deleteRes.SuccessIDs = append(deleteRes.SuccessIDs, appId)
		successApps = append(successApps, appInfo.AppName)
	}
	logmgmt.BatchOperationLog("batch delete app", successApps)
	if len(deleteRes.FailedInfos) != 0 {
		return common.RespMsg{Status: common.ErrorDeleteApp, Msg: "", Data: deleteRes}
	}

	hwlog.RunLog.Info("app db delete success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

// listAppInstancesById get deployed apps' list
func listAppInstancesById(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start list app instances by id")

	var appId uint64
	if err := msg.ParseContent(&appId); err != nil {
		hwlog.RunLog.Errorf("list app instances failed, parse param failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse param failed", Data: nil}
	}
	if checkResult := appchecker.IdChecker().Check(appId); !checkResult.Result {
		hwlog.RunLog.Errorf("list app instances failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}

	appInstances, err := AppRepositoryInstance().listAppInstancesById(appId)
	if err != nil {
		hwlog.RunLog.Error("list app instances db failed")
		return common.RespMsg{Status: common.ErrorListAppInstancesByID, Msg: "list app instances db failed", Data: nil}
	}
	appInstanceResp, err := getAppInstanceRespFromAppInstances(appInstances)
	if err != nil {
		hwlog.RunLog.Error("get app instance response from app instances failed")
		return common.RespMsg{Status: common.ErrorListAppInstancesByID,
			Msg: "get app instance response from app instances failed", Data: nil}
	}
	resp := &ListAppInstancesResp{
		AppInstances: appInstanceResp,
		Total:        int64(len(appInstanceResp)),
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

func getAppInstanceRespFromAppInstances(appInstances []AppInstance) ([]AppInstanceResp, error) {
	var appInstanceResp []AppInstanceResp
	for _, instance := range appInstances {
		nodeStatus, err := getNodeStatus(instance.NodeUniqueName)
		if err != nil {
			hwlog.RunLog.Warnf("get node [%s] status error: %v", instance.NodeUniqueName, err)
			continue
		}
		podStatus := appStatusService.getPodStatusFromCache(instance.PodName, nodeStatus)
		containerInfos, err := appStatusService.getContainerInfos(instance, nodeStatus)
		if err != nil {
			hwlog.RunLog.Warnf("get app id [%d] of node [%d] container infos error: %v",
				instance.AppID, instance.NodeID, err)
			continue
		}

		nodeInfos, err := getNodeGroupInfos([]uint64{instance.NodeGroupID})
		if err != nil || len(nodeInfos) != 1 {
			hwlog.RunLog.Warnf("get node group [%d] name failed, node manager error", instance.NodeGroupID)
			continue
		}

		createdAt := instance.CreatedAt.Format(common.TimeFormat)
		resp := AppInstanceResp{
			AppID:   instance.AppID,
			AppName: instance.AppName,
			NodeGroupInfo: types.NodeGroupInfo{
				NodeGroupID:   instance.NodeGroupID,
				NodeGroupName: nodeInfos[0].NodeGroupName,
			},
			NodeID:        instance.NodeID,
			NodeName:      instance.NodeName,
			NodeStatus:    nodeStatus,
			AppStatus:     podStatus,
			CreatedAt:     createdAt,
			ContainerInfo: containerInfos,
		}
		appInstanceResp = append(appInstanceResp, resp)
	}
	return appInstanceResp, nil
}

// listAppInstancesByNode get deployed apps' list of a certain node
func listAppInstancesByNode(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start list app instances by node id")

	var nodeId uint64
	if err := msg.ParseContent(&nodeId); err != nil {
		hwlog.RunLog.Errorf("list app instances by node id failed, parse param failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse param failed", Data: nil}
	}
	if checkResult := appchecker.IdChecker().Check(nodeId); !checkResult.Result {
		hwlog.RunLog.Errorf("list app instances by node failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}

	appInstances, err := AppRepositoryInstance().listAppInstancesByNode(nodeId)
	if err != nil {
		hwlog.RunLog.Error("list app instances by node failed, db failed")
		return common.RespMsg{Status: common.ErrorListAppInstancesByNode, Msg: "list app instances by node db failed"}
	}
	appList, err := getAppInstanceRespFromAppInstances(appInstances)
	if err != nil {
		hwlog.RunLog.Error("get app instance of node response from app instances failed")
		return common.RespMsg{Status: common.ErrorListAppInstancesByNode,
			Msg: "get app instance of node response from app instances failed", Data: nil}
	}
	resp := &ListAppInstancesResp{
		AppInstances: appList,
		Total:        int64(len(appList)),
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

func listAppInstances(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start to list all app instances")

	var req types.ListReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("list all app instances failed: parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}
	if checkResult := util.NewPaginationQueryChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("list all app instances failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}

	appInstances, err := AppRepositoryInstance().listAppInstances(req.PageNum, req.PageSize, req.Name)
	if err != nil {
		hwlog.RunLog.Error("list all app instances failed: db failed")
		return common.RespMsg{Status: common.ErrorListAppInstances, Msg: "list all app instances failed: db failed"}
	}
	appInstanceResp, err := getAppInstanceRespFromAppInstances(appInstances)
	if err != nil {
		hwlog.RunLog.Error("list all app instances from app instances failed")
		return common.RespMsg{Status: common.ErrorListAppInstances, Msg: "list all app instances failed", Data: nil}
	}
	total, err := AppRepositoryInstance().countListAppInstances(req.Name)
	if err != nil {
		hwlog.RunLog.Error("count all app instances list failed")
		return common.RespMsg{Status: common.ErrorListAppInstances, Msg: "count app instances list failed", Data: nil}
	}
	resp := &ListAppInstancesResp{
		AppInstances: appInstanceResp,
		Total:        total,
	}

	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}
