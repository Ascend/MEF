// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager service
package appmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/appmanager/appchecker"
	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/types"
)

// createApp Create application
func createApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start create app")
	var req CreateAppReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
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
	if total+1 >= MaxApp {
		hwlog.RunLog.Error("app number is enough, can not create")
		return common.RespMsg{Status: common.ErrorCheckAppMrgSize, Msg: "app number is enough, can not create"}
	}
	app, err := req.toDb()
	if err != nil {
		hwlog.RunLog.Error("create app requst conver to db failed ")
		return common.RespMsg{Status: common.ErrorAppParamConvertDb, Msg: "get appInfo failed", Data: nil}
	}
	if err = AppRepositoryInstance().createApp(app); err != nil {
		if strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
			hwlog.RunLog.Error("app name is duplicate")
			return common.RespMsg{Status: common.ErrorAppMrgDuplicate, Msg: "app name is duplicate", Data: nil}
		}
		hwlog.RunLog.Error("app db create failed")
		return common.RespMsg{Status: common.ErrorCreateApp, Msg: "app db create failed", Data: nil}
	}
	hwlog.RunLog.Info("app db create success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: app.ID}
}

// queryApp app info
func queryApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start query app info")

	appId, ok := input.(uint64)
	if !ok {
		hwlog.RunLog.Error("query app info failed: para type not valid")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "query app request convert error", Data: nil}
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

	var resp AppReturnInfo
	resp.AppID = appInfo.ID
	resp.AppName = appInfo.AppName
	resp.Description = appInfo.Description

	if err = json.Unmarshal([]byte(appInfo.Containers), &resp.Containers); err != nil {
		hwlog.RunLog.Error("unmarshal containers info failed")
		return common.RespMsg{Status: common.ErrorUnmarshalContainer, Msg: "unmarshal containers info failed", Data: nil}
	}

	hwlog.RunLog.Info("query app success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

// listAppInfo get appInfo list
func listAppInfo(input interface{}) common.RespMsg {
	hwlog.RunLog.Infof("start list app infos")
	req, ok := input.(types.ListReq)
	if !ok {
		hwlog.RunLog.Error("get apps Infos list failed: para type is invalid")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "", Data: nil}
	}

	apps, err := getListReturnInfo(req)
	if err == gorm.ErrRecordNotFound {
		hwlog.RunLog.Info("dont have any apps")
		return common.RespMsg{Status: common.Success, Msg: "dont have any apps", Data: nil}
	}
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
		nodeGroupInfos, err := AppRepositoryInstance().getNodeGroupInfosByAppID(app.ID)
		if err != nil {
			hwlog.RunLog.Error("get node group name failed when list")
			return nil, err
		}
		appReturnInfo := AppReturnInfo{
			AppID:          app.ID,
			AppName:        app.AppName,
			Description:    app.Description,
			CreatedAt:      app.CreatedAt.Format(common.TimeFormat),
			ModifiedAt:     app.UpdatedAt.Format(common.TimeFormat),
			NodeGroupInfos: nodeGroupInfos,
			Containers:     containers,
		}
		appReturnInfos = append(appReturnInfos, appReturnInfo)
	}
	return &ListReturnInfo{
		AppInfo: appReturnInfos,
	}, nil
}

// deployApp deploy application on node group
func deployApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start deploy app")
	var req DeployAppReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}
	checker := deployParaChecker{req: &req}
	if err := checker.Check(); err != nil {
		hwlog.RunLog.Errorf("deploy app para check failed: %s", err.Error())
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: err.Error(), Data: nil}
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
	var deployRes types.BatchResp
	for _, nodeGroupId := range req.NodeGroupIds {
		daemonSet, err := initDaemonSet(appInfo, nodeGroupId)
		if err != nil {
			hwlog.RunLog.Errorf("init daemonSet app [%s] on node group id [%d] failed: %s",
				appInfo.AppName, nodeGroupId, err.Error())
			deployRes.FailedIDs = append(deployRes.FailedIDs, nodeGroupId)
			continue
		}
		daemonSet, err = kubeclient.GetKubeClient().CreateDaemonSet(daemonSet)
		if err != nil {
			hwlog.RunLog.Errorf("app daemonSet create failed: %s", err.Error())
			deployRes.FailedIDs = append(deployRes.FailedIDs, nodeGroupId)
			continue
		}
		deployRes.SuccessIDs = append(deployRes.SuccessIDs, nodeGroupId)
	}
	if len(deployRes.FailedIDs) != 0 {
		return common.RespMsg{Status: common.ErrorDeployApp, Msg: "", Data: deployRes}
	}
	hwlog.RunLog.Info("all app daemonSets create success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

// unDeployApp deploy application on node group
func unDeployApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start unDeploy app")
	var req UndeployAppReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}
	checker := undeployParaParser{req: &req}
	if err := checker.Parse(); err != nil {
		hwlog.RunLog.Errorf("undeploy app para check failed: %s", err.Error())
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: err.Error(), Data: nil}
	}
	if len(req.NodeGroupIds) == 0 {
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: "do not have any group can be undeploy", Data: nil}
	}
	appInfo, err := AppRepositoryInstance().getAppInfoById(req.AppID)
	if err != nil {
		hwlog.RunLog.Error("get app info error, undeploy app failed")
		return common.RespMsg{Status: common.ErrorUnDeployApp, Msg: "get app info error, undeploy app failed"}
	}
	var unDeployRes types.BatchResp
	for _, nodeGroupId := range req.NodeGroupIds {
		daemonSetName := formatDaemonSetName(appInfo.AppName, nodeGroupId)
		if err = kubeclient.GetKubeClient().DeleteDaemonSet(daemonSetName); err != nil {
			hwlog.RunLog.Errorf("undeploy app [%s] on node group id [%d] failed: %s",
				appInfo.AppName, nodeGroupId, err.Error())
			unDeployRes.FailedIDs = append(unDeployRes.FailedIDs, nodeGroupId)
			continue
		}
		unDeployRes.SuccessIDs = append(unDeployRes.SuccessIDs, nodeGroupId)
	}
	if len(unDeployRes.FailedIDs) != 0 {
		return common.RespMsg{Status: common.ErrorUnDeployApp, Msg: "undeploy app failed", Data: unDeployRes}
	}
	hwlog.RunLog.Info("undeploy app on node group success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
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
func updateApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start update app")
	var req UpdateAppReq
	var err error
	if err = common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}

	if checkResult := appchecker.NewUpdateAppChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("app update para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}
	if err := NewAppSupplementalChecker(req.CreateAppReq).Check(); err != nil {
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

	if err = AppRepositoryInstance().updateApp(appInfo.ID, "containers", appInfo.Containers); err != nil {
		if strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
			hwlog.RunLog.Error("update app to db failed")
			return common.RespMsg{Status: common.ErrorAppMrgDuplicate, Msg: "update app to db failed", Data: nil}
		}
		hwlog.RunLog.Error("update app to db failed")
		return common.RespMsg{Status: common.ErrorUpdateApp, Msg: "update app to db failed", Data: nil}
	}

	nodeGroups, err := AppRepositoryInstance().queryNodeGroup(req.AppID)
	if err != nil {
		hwlog.RunLog.Error("get node group failed ")
		return common.RespMsg{Status: common.ErrorUpdateApp, Msg: "get node group failed", Data: nil}
	}

	if err = updateNodeGroupDaemonSet(appInfo, nodeGroups); err != nil {
		hwlog.RunLog.Errorf("update node group daemon set failed: %s", err.Error())
		return common.RespMsg{Status: common.ErrorUpdateApp, Msg: "update node group daemon set failed", Data: nil}
	}

	hwlog.RunLog.Info("app daemonSet update success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

// deleteApp delete application by appName
func deleteApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start delete app")
	var req DeleteAppReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}
	if checkResult := appchecker.NewDeleteAppChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("app delete para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}
	var deleteRes types.BatchResp
	for _, appId := range req.AppIDs {
		if rowsAffected, err := AppRepositoryInstance().deleteAppById(appId); err != nil || rowsAffected != 1 {
			hwlog.RunLog.Errorf("app db delete failed: %v", err)
			deleteRes.FailedIDs = append(deleteRes.FailedIDs, appId)
			continue
		}
		deleteRes.SuccessIDs = append(deleteRes.SuccessIDs, appId)
	}
	if len(deleteRes.FailedIDs) != 0 {
		return common.RespMsg{Status: common.ErrorDeleteApp, Msg: "", Data: deleteRes}
	}
	hwlog.RunLog.Info("app db delete success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

// listAppInstancesById get deployed apps' list
func listAppInstancesById(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start list app instances by id")
	var appId uint64
	appId, ok := input.(uint64)
	if !ok {
		hwlog.RunLog.Error("list app instances failed, param type is not integer")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "param type is not integer", Data: nil}
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
	return common.RespMsg{Status: common.Success, Msg: "", Data: appInstanceResp}
}

func getAppInstanceRespFromAppInstances(appInstances []AppInstance) ([]AppInstanceResp, error) {
	var appInstanceResp []AppInstanceResp
	for _, instance := range appInstances {
		nodeStatus, err := getNodeStatus(instance.NodeUniqueName)
		if err != nil {
			hwlog.RunLog.Errorf("get node [%s] status error: %v", instance.NodeUniqueName, err)
			return nil, err
		}
		podStatus := appStatusService.getPodStatusFromCache(instance.PodName, nodeStatus)
		containerInfos, err := appStatusService.getContainerInfos(instance, nodeStatus)
		if err != nil {
			hwlog.RunLog.Errorf("get app id [%d] of node [%d] container infos error: %v",
				instance.AppID, instance.NodeID, err)
			return nil, err
		}
		nodeGroupName, err := AppRepositoryInstance().getNodeGroupName(instance.AppID, instance.NodeGroupID)
		if err != nil {
			hwlog.RunLog.Errorf("get app id [%d] node group [%d] name failed, db error",
				instance.AppID, instance.NodeGroupID)
			return nil, err
		}
		createdAt := instance.CreatedAt.Format(common.TimeFormat)
		resp := AppInstanceResp{
			AppName: instance.AppName,
			NodeGroupInfo: types.NodeGroupInfo{
				NodeGroupID:   instance.NodeGroupID,
				NodeGroupName: nodeGroupName,
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

func getNodeStatus(nodeUniqueName string) (string, error) {
	if nodeUniqueName == "" {
		hwlog.RunLog.Warn("app instance node name is empty, pod is in pending phase")
		return "", nil
	}
	router := common.Router{
		Source:      common.AppManagerName,
		Destination: common.NodeManagerName,
		Option:      common.Inner,
		Resource:    common.NodeStatus,
	}
	req := types.InnerGetNodeStatusReq{
		UniqueName: nodeUniqueName,
	}
	resp := common.SendSyncMessageByRestful(req, &router)
	if resp.Status == "" {
		return "", fmt.Errorf("get info from other module error, %v", resp.Msg)
	}
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return "", errors.New("marshal internal response error")
	}
	var node types.InnerGetNodeStatusResp
	if err = json.Unmarshal(data, &node); err != nil {
		return "", errors.New("unmarshal internal response error")
	}
	return node.NodeStatus, nil
}

// listAppInstancesByNode get deployed apps' list of a certain node
func listAppInstancesByNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start list app instances by node id")
	var nodeId uint64
	nodeId, ok := input.(uint64)
	if !ok {
		hwlog.RunLog.Error("list app instances by node id failed, param type is not integer")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "param type is not integer", Data: nil}
	}
	appInstances, err := AppRepositoryInstance().listAppInstancesByNode(nodeId)
	if err != nil {
		hwlog.RunLog.Error("list app instances by node failed, db failed")
		return common.RespMsg{Status: common.ErrorListAppInstancesByNode, Msg: "list app instances by node db failed"}
	}
	appList, err := getAppInstanceOfNodeRespFromAppInstances(appInstances)
	if err != nil {
		hwlog.RunLog.Error("get app instance of node response from app instances failed")
		return common.RespMsg{Status: common.ErrorListAppInstancesByNode,
			Msg: "get app instance of node response from app instances failed", Data: nil}
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: appList}
}

func getAppInstanceOfNodeRespFromAppInstances(appInstances []AppInstance) ([]AppInstanceOfNodeResp, error) {
	var appList []AppInstanceOfNodeResp
	for _, instance := range appInstances {
		appInfo, err := AppRepositoryInstance().getAppInfoByName(instance.AppName)
		if err != nil {
			hwlog.RunLog.Errorf("get app info by name [%s] db error", instance.AppName)
			return nil, err
		}
		nodeGroupName, err := AppRepositoryInstance().getNodeGroupName(instance.AppID, instance.NodeGroupID)
		if err != nil {
			hwlog.RunLog.Errorf("get app id [%d] node group [%d] name failed, db error",
				instance.AppID, instance.NodeGroupID)
			return nil, err
		}
		nodeStatus, err := getNodeStatus(instance.NodeUniqueName)
		if err != nil {
			hwlog.RunLog.Errorf("get node [%s] status error: %v", instance.NodeUniqueName, err)
			return nil, err
		}
		status := appStatusService.getPodStatusFromCache(instance.PodName, nodeStatus)
		createdAt := instance.CreatedAt.Format(common.TimeFormat)
		changedAt := instance.UpdatedAt.Format(common.TimeFormat)
		instanceResp := AppInstanceOfNodeResp{
			AppName:     instance.AppName,
			AppStatus:   status,
			Description: appInfo.Description,
			CreatedAt:   createdAt,
			ChangedAt:   changedAt,
			NodeGroupInfo: types.NodeGroupInfo{
				NodeGroupID:   instance.NodeGroupID,
				NodeGroupName: nodeGroupName,
			},
		}
		appList = append(appList, instanceResp)
	}
	return appList, nil
}

func listAppInstances(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start to list all app instances")
	req, ok := input.(types.ListReq)
	if !ok {
		hwlog.RunLog.Error("list all app instances failed: para type is invalid")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "para type is invalid", Data: nil}
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

func getAppInstanceCountByNodeGroup(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start to get appInstance count")
	req, ok := input.([]uint64)
	if !ok {
		hwlog.RunLog.Error("failed to convert param")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "failed to convert param"}
	}
	appInstanceCount := make(map[uint64]int64)
	for _, groupId := range req {
		count, err := AppRepositoryInstance().countDeployedAppByGroupID(groupId)
		if err != nil {
			hwlog.RunLog.Error("failed to count appInstance by node group")
			return common.RespMsg{Status: common.ErrorGetAppInstanceCountByNodeGroup, Msg: ""}
		}
		appInstanceCount[groupId] = count
	}
	hwlog.RunLog.Info("get appInstance count success")
	return common.RespMsg{Status: common.Success, Data: appInstanceCount}
}
