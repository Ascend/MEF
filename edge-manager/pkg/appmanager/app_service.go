// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager service
package appmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"k8s.io/api/apps/v1"

	"edge-manager/pkg/appmanager/appchecker"
	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/types"
	"edge-manager/pkg/util"
	"huawei.com/mindxedge/base/common"
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

	id, err := AppRepositoryInstance().createAppAndUpdateCm(&req)
	if err != nil {
		if err.Error() == common.ErrDbUniqueFailed {
			return common.RespMsg{Status: common.ErrorAppMrgDuplicate, Msg: "app name is duplicate", Data: nil}
		}
		return common.RespMsg{Status: common.ErrorCreateApp, Msg: err.Error(), Data: nil}
	}

	hwlog.RunLog.Info("app db create success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: id}
}

// queryApp app info
func queryApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start query app info")

	appId, ok := input.(uint64)
	if !ok {
		hwlog.RunLog.Error("query app info failed: para type not valid")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "query app request convert error", Data: nil}
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

	nodeGroupInfos, err := AppRepositoryInstance().getNodeGroupInfosByAppID(appId)
	if err != nil {
		hwlog.RunLog.Errorf("query app info failed, db error")
		return common.RespMsg{Status: common.ErrorQueryApp, Msg: "query app info failed", Data: nil}
	}
	resp := AppReturnInfo{
		AppID:          appInfo.ID,
		AppName:        appInfo.AppName,
		Description:    appInfo.Description,
		CreatedAt:      appInfo.CreatedAt.Format(common.TimeFormat),
		ModifiedAt:     appInfo.UpdatedAt.Format(common.TimeFormat),
		NodeGroupInfos: nodeGroupInfos,
	}

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
	deployRes := deployAppToNodeGroups(appInfo, req.NodeGroupIds)

	if len(deployRes.FailedInfos) != 0 {
		return common.RespMsg{Status: common.ErrorDeployApp, Msg: "", Data: deployRes}
	}

	hwlog.RunLog.Info("all app daemonSets create success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func deployAppToNodeGroups(appInfo *AppInfo, NodeGroupIds []uint64) types.BatchResp {
	var deployRes types.BatchResp
	failedMap := make(map[string]string)
	deployRes.FailedInfos = failedMap

	deployedNode := make(map[uint64]int)
	for _, nodeGroupId := range NodeGroupIds {
		_, err := getNodeGroupInfos([]uint64{nodeGroupId})
		if err != nil {
			errInfo := fmt.Sprintf("init daemonSet app [%s] on node group id [%d] failed: group id no exist",
				appInfo.AppName, nodeGroupId)
			hwlog.RunLog.Error(errInfo)
			failedMap[strconv.Itoa(int(nodeGroupId))] = errInfo
			continue
		}
		daemonSet, err := initDaemonSet(appInfo, nodeGroupId)
		if err != nil {
			errInfo := fmt.Sprintf("init daemonSet app [%s] on node group id [%d] failed: %s",
				appInfo.AppName, nodeGroupId, err.Error())
			hwlog.RunLog.Error(errInfo)
			failedMap[strconv.Itoa(int(nodeGroupId))] = errInfo
			continue
		}
		if err := checkNodeGroupRes(nodeGroupId, daemonSet, deployedNode); err != nil {
			errInfo := fmt.Sprintf("check app [%s] resources on node group id [%d] failed: %s",
				appInfo.AppName, nodeGroupId, err.Error())
			hwlog.RunLog.Error(errInfo)
			failedMap[strconv.Itoa(int(nodeGroupId))] = errInfo
			continue
		}
		daemonSet, err = kubeclient.GetKubeClient().CreateDaemonSet(daemonSet)
		if err != nil {
			errInfo := fmt.Sprintf("app daemonSet create failed: %s", err.Error())
			hwlog.RunLog.Error(errInfo)
			failedMap[strconv.Itoa(int(nodeGroupId))] = errInfo
			continue
		}
		deployRes.SuccessIDs = append(deployRes.SuccessIDs, nodeGroupId)
	}
	return deployRes
}

func checkNodeGroupRes(nodeGroupId uint64, daemonSet *v1.DaemonSet, deployedNode map[uint64]int) error {
	if deployedNode == nil {
		return errors.New("nil map error")
	}
	var duplicatedCount int
	nodeIDs, err := getNodesByNodeGroup(nodeGroupId)
	if err != nil {
		return fmt.Errorf("get nodes by group id [%d] failed", nodeGroupId)
	}
	for _, nodeID := range nodeIDs {
		count := deployedNode[nodeID]
		count++
		deployedNode[nodeID] = count
		if count > duplicatedCount {
			duplicatedCount = count
		}
	}
	return checkNodeGroupResWithDuplicatedNode(nodeGroupId, daemonSet, duplicatedCount)
}

// unDeployApp deploy application on node group
func unDeployApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start unDeploy app")

	var req UndeployAppReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
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

	var unDeployRes types.BatchResp
	failedMap := make(map[string]string)
	unDeployRes.FailedInfos = failedMap
	for _, nodeGroupId := range req.NodeGroupIds {
		daemonSetName := formatDaemonSetName(appInfo.AppName, nodeGroupId)
		if err = kubeclient.GetKubeClient().DeleteDaemonSet(daemonSetName); err != nil {
			errInfo := fmt.Sprintf("undeploy app [%s] on node group id [%d] failed: %s",
				appInfo.AppName, nodeGroupId, err.Error())
			hwlog.RunLog.Error(errInfo)
			failedMap[strconv.Itoa(int(nodeGroupId))] = errInfo
			continue
		}
		unDeployRes.SuccessIDs = append(unDeployRes.SuccessIDs, nodeGroupId)
	}
	if len(unDeployRes.FailedInfos) != 0 {
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
	failedMap := make(map[string]string)
	deleteRes.FailedInfos = failedMap

	for _, appId := range req.AppIDs {
		if err := AppRepositoryInstance().isAppReferenced(appId); err != nil {
			failedMap[strconv.Itoa(int(appId))] = err.Error()
			continue
		}

		if err := AppRepositoryInstance().deleteSingleApp(appId); err != nil {
			failedMap[strconv.Itoa(int(appId))] = err.Error()
			continue
		}

		deleteRes.SuccessIDs = append(deleteRes.SuccessIDs, appId)
	}
	if len(deleteRes.FailedInfos) != 0 {
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
		nodeGroupName, err := AppRepositoryInstance().getNodeGroupName(instance.AppID, instance.NodeGroupID)
		if err != nil {
			hwlog.RunLog.Warnf("get app id [%d] node group [%d] name failed, db error",
				instance.AppID, instance.NodeGroupID)
			continue
		}
		createdAt := instance.CreatedAt.Format(common.TimeFormat)
		resp := AppInstanceResp{
			AppID:   instance.AppID,
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

// listAppInstancesByNode get deployed apps' list of a certain node
func listAppInstancesByNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start list app instances by node id")

	var nodeId uint64
	nodeId, ok := input.(uint64)
	if !ok {
		hwlog.RunLog.Error("list app instances by node id failed, param type is not integer")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "param type is not integer", Data: nil}
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

func listAppInstances(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start to list all app instances")

	req, ok := input.(types.ListReq)
	if !ok {
		hwlog.RunLog.Error("list all app instances failed: para type is invalid")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "para type is invalid", Data: nil}
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
