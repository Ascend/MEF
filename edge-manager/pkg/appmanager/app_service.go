// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager service
package appmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/types"
	"edge-manager/pkg/util"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

// createApp Create application
func createApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start create app")
	var req CreateAppReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	checker := appCreatParaChecker{req: &req}
	if err := checker.Check(); err != nil {
		hwlog.RunLog.Errorf("app create para check failed: %s", err.Error())
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	total, err := GetTableCount(AppInfo{})
	if err != nil {
		hwlog.RunLog.Error("get app table num failed")
		return common.RespMsg{Status: "", Msg: "get app table num failed", Data: nil}
	}
	if total >= MaxApp {
		hwlog.RunLog.Error("app number is enough, can not create")
		return common.RespMsg{Status: "", Msg: "app number is enough, can not create", Data: nil}
	}
	app, err := req.toDb()
	if err != nil {
		hwlog.RunLog.Error("get appInfo failed ")
		return common.RespMsg{Status: "", Msg: "get appInfo failed", Data: nil}
	}
	if err = AppRepositoryInstance().createApp(app); err != nil {
		if strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
			hwlog.RunLog.Error("app name is duplicate")
			return common.RespMsg{Status: "", Msg: "app name is duplicate", Data: nil}
		}
		hwlog.RunLog.Error("app db create failed")
		return common.RespMsg{Status: "", Msg: "app db create failed", Data: nil}
	}
	appInfo, err := AppRepositoryInstance().getAppInfoByName(app.AppName)
	if err != nil {
		hwlog.RunLog.Error("get app id failed when create")
		return common.RespMsg{Status: "", Msg: "get app id failed when create", Data: nil}
	}
	createReturnInfo := CreateReturnInfo{
		AppId: appInfo.ID,
	}
	hwlog.RunLog.Info("app db create success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: createReturnInfo}
}

// queryApp app info
func queryApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start query app info")
	appId, ok := input.(uint64)
	if !ok {
		hwlog.RunLog.Error("query app info failed")
		return common.RespMsg{Status: "", Msg: "query app info failed", Data: nil}
	}
	appInfo, err := AppRepositoryInstance().queryApp(appId)
	if err != nil {
		hwlog.RunLog.Error("query app info failed")
		return common.RespMsg{Status: "", Msg: "query app info failed", Data: nil}
	}

	var resp AppReturnInfo
	resp.AppId = appInfo.ID
	resp.AppName = appInfo.AppName
	resp.Description = appInfo.Description

	if err = json.Unmarshal([]byte(appInfo.Containers), &resp.Containers); err != nil {
		hwlog.RunLog.Error("unmarshal containers info failed")
		return common.RespMsg{Status: "", Msg: "unmarshal containers info failed", Data: nil}
	}

	hwlog.RunLog.Info("query app success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

// listAppInfo get appInfo list
func listAppInfo(input interface{}) common.RespMsg {
	hwlog.RunLog.Infof("start list app infos")
	req, ok := input.(util.ListReq)
	if !ok {
		hwlog.RunLog.Error("get apps Infos list failed: para type is invalid")
		return common.RespMsg{Status: "", Msg: "list app info error", Data: nil}
	}

	apps, err := AppRepositoryInstance().listAppsInfo(req.PageNum, req.PageSize, req.Name)
	if err == gorm.ErrRecordNotFound {
		hwlog.RunLog.Info("dont have any apps")
		return common.RespMsg{Status: common.Success, Msg: "dont have any apps", Data: nil}
	}
	if err != nil {
		hwlog.RunLog.Error("get apps Infos list failed")
		return common.RespMsg{Status: "", Msg: "get apps Infos list failed", Data: nil}
	}

	apps.Total, err = AppRepositoryInstance().countListAppsInfo(req.Name)
	if err != nil {
		hwlog.RunLog.Error("count apps Infos list failed")
		return common.RespMsg{Status: "", Msg: "count apps Infos list failed", Data: nil}
	}
	apps.Deployed, apps.UnDeployed, err = AppRepositoryInstance().countDeployedApp()
	if err != nil {
		hwlog.RunLog.Error("count deployed app failed")
		return common.RespMsg{Status: "", Msg: "count apps Infos list failed", Data: nil}
	}
	hwlog.RunLog.Info("list apps Infos success")
	return common.RespMsg{Status: common.Success, Msg: "list apps Infos success", Data: apps}
}

// deployApp deploy application on node group
func deployApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start deploy app")
	var req DeployAppReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	appInfo, err := AppRepositoryInstance().getAppInfoById(req.AppId)
	if err != nil {
		hwlog.RunLog.Error("get app information failed")
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	for _, nodeGroupId := range req.NodeGroupIds {
		daemonSet, err := initDaemonSet(appInfo, nodeGroupId)
		if err != nil {
			hwlog.RunLog.Errorf("deploy app [%s] on node group id [%d] failed: %s",
				appInfo.AppName, nodeGroupId, err.Error())
			return common.RespMsg{Status: "", Msg: "app daemonSet init failed", Data: nil}
		}
		daemonSet, err = kubeclient.GetKubeClient().CreateDaemonSet(daemonSet)
		if err != nil {
			hwlog.RunLog.Errorf("app daemonSet create failed: %s", err.Error())
			return common.RespMsg{Status: "", Msg: "app daemonSet create failed", Data: nil}
		}
		hwlog.RunLog.Infof("deploy app [%s] on node group id [%d] success",
			appInfo.AppName, nodeGroupId)
	}

	hwlog.RunLog.Info("all app daemonSets create success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

// unDeployApp deploy application on node group
func unDeployApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start unDeploy app")
	var req UndeployAppReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	appInfo, err := AppRepositoryInstance().getAppInfoById(req.AppId)
	if err != nil {
		hwlog.RunLog.Error("get app information failed")
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	for _, nodeGroupId := range req.NodeGroupIds {
		daemonSetName := appInfo.AppName + "-" + strconv.FormatInt(nodeGroupId, DecimalScale)
		if err = kubeclient.GetKubeClient().DeleteDaemonSet(daemonSetName); err != nil {
			hwlog.RunLog.Errorf("undeploy app [%s] on node group id [%d] failed: %s",
				appInfo.AppName, nodeGroupId, err.Error())
			return common.RespMsg{Status: "", Msg: "undeploy app failed", Data: nil}
		}
		hwlog.RunLog.Infof("undeploy app [%s] on node group id [%d] success",
			appInfo.AppName, nodeGroupId)
	}

	hwlog.RunLog.Info("undeploy app on node group success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func updateNodeGroupDaemonSet(appInfo *AppInfo, nodeGroups []NodeGroupInfo) error {
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

// updateApp update application
func updateApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start update app")
	var req UpdateAppReq
	var err error
	if err = common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	checker := appCreatParaChecker{req: &req.CreateAppReq}
	if err = checker.Check(); err != nil {
		hwlog.RunLog.Errorf("app update para check failed: %s", err.Error())
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	appInfo, err := req.toDb()
	if err != nil {
		hwlog.RunLog.Error("get app info failed ")
		return common.RespMsg{Status: "", Msg: "get app info failed", Data: nil}
	}

	if err = AppRepositoryInstance().updateApp(appInfo.ID, "containers", appInfo.Containers); err != nil {
		if strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
			hwlog.RunLog.Error("update app to db failed")
			return common.RespMsg{Status: "", Msg: "update app to db failed", Data: nil}
		}
		hwlog.RunLog.Error("update app to db failed")
		return common.RespMsg{Status: "", Msg: "update app to db failed", Data: nil}
	}

	nodeGroups, err := AppRepositoryInstance().queryNodeGroup(req.AppId)
	if err != nil {
		hwlog.RunLog.Error("get node group failed ")
		return common.RespMsg{Status: "", Msg: "get node group failed", Data: nil}
	}

	if err = updateNodeGroupDaemonSet(appInfo, nodeGroups); err != nil {
		hwlog.RunLog.Errorf("update node group daemon set failed: %s", err.Error())
		return common.RespMsg{Status: "", Msg: "update node group daemon set failed", Data: nil}
	}

	hwlog.RunLog.Info("app daemonSet update success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

// deleteApp delete application by appName
func deleteApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start delete app")
	var req DeleteAppReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	for _, appId := range req.AppIds {
		if err := AppRepositoryInstance().deleteAppById(appId); err != nil {
			hwlog.RunLog.Error("app db delete failed")
			return common.RespMsg{Status: "", Msg: "app db delete failed", Data: nil}
		}
	}

	hwlog.RunLog.Info("app db delete success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

// listAppInstances get deployed apps' list
func listAppInstances(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start list app instances by id")
	var appId uint64
	appId, ok := input.(uint64)
	if !ok {
		hwlog.RunLog.Error("list app instances failed, param type is not integer")
		return common.RespMsg{Status: "", Msg: "param type is not integer", Data: nil}
	}
	appInstances, err := AppRepositoryInstance().listAppInstances(appId)
	if err != nil {
		hwlog.RunLog.Error("list app instances db failed")
		return common.RespMsg{Status: "", Msg: "list app instances db failed", Data: nil}
	}
	appInstanceResp, err := getAppInstanceRespFromAppInstances(appInstances)
	if err != nil {
		hwlog.RunLog.Error("get app instance response from app instances failed")
		return common.RespMsg{Status: "", Msg: "get app instance response from app instances failed", Data: nil}
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: appInstanceResp}
}

func getAppInstanceRespFromAppInstances(appInstances []AppInstance) ([]AppInstanceResp, error) {
	var appInstanceResp []AppInstanceResp
	for _, instance := range appInstances {
		nodeName := instance.NodeName
		nodeStatus, err := getNodeStatus(nodeName)
		if err != nil {
			hwlog.RunLog.Errorf("get node status error: %v", err)
			return nil, err
		}
		podStatus := appStatusService.getPodStatusFromCache(instance.PodName)
		containerInfos, err := appStatusService.getContainerInfos(instance)
		if err != nil {
			hwlog.RunLog.Errorf("get container infos error: %v", err)
			return nil, err
		}
		createdAt := instance.CreatedAt
		if err != nil {
			hwlog.RunLog.Errorf("parse db time to standard format error: %v", err)
			return nil, err
		}
		resp := AppInstanceResp{
			AppName: instance.AppName,
			NodeGroupInfo: NodeGroupInfo{NodeGroupID: instance.NodeGroupID,
				NodeGroupName: instance.NodeGroupName},
			NodeId:        instance.NodeID,
			NodeName:      nodeName,
			NodeStatus:    nodeStatus,
			AppStatus:     podStatus,
			CreatedAt:     createdAt.Format(common.TimeFormat),
			ContainerInfo: containerInfos,
		}
		appInstanceResp = append(appInstanceResp, resp)
	}
	return appInstanceResp, nil
}

func getNodeStatus(nodeName string) (string, error) {
	if nodeName == "" {
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
		UniqueName: nodeName,
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
	var nodeId int64
	nodeId, ok := input.(int64)
	if !ok {
		hwlog.RunLog.Error("list app instances by node id failed, param type is not integer")
		return common.RespMsg{Status: "", Msg: "param type is not integer", Data: nil}
	}
	appInstances, err := AppRepositoryInstance().listAppInstancesByNode(nodeId)
	if err != nil {
		hwlog.RunLog.Error("list app instances by node failed, db failed")
		return common.RespMsg{Status: "", Msg: "list app instances by node db failed", Data: nil}
	}
	appList, err := getAppInstanceOfNodeRespFromAppInstances(appInstances)
	if err != nil {
		hwlog.RunLog.Error("get app instance of node response from app instances failed")
		return common.RespMsg{Status: "", Msg: "get app instance of node response from app instances failed", Data: nil}
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: appList}
}

func getAppInstanceOfNodeRespFromAppInstances(appInstances []AppInstance) ([]AppInstanceOfNodeResp, error) {
	var appList []AppInstanceOfNodeResp
	for _, instance := range appInstances {
		appInfo, err := AppRepositoryInstance().getAppInfoByName(instance.AppName)
		if err != nil {
			hwlog.RunLog.Errorf("get app info by name db error: %v", err)
			return nil, err
		}
		status := appStatusService.getPodStatusFromCache(instance.PodName)
		createdAt := instance.CreatedAt
		if err != nil {
			hwlog.RunLog.Errorf("parse db time to standard format error: %v", err)
			return nil, err
		}
		changedAt := instance.UpdatedAt
		if err != nil {
			hwlog.RunLog.Errorf("parse db time to standard format error: %v", err)
			return nil, err
		}
		instanceResp := AppInstanceOfNodeResp{
			AppName:       instance.AppName,
			AppStatus:     status,
			Description:   appInfo.Description,
			CreatedAt:     createdAt.Format(common.TimeFormat),
			ChangedAt:     changedAt.Format(common.TimeFormat),
			NodeGroupName: instance.NodeGroupName,
			NodeGroupID:   instance.NodeGroupID,
		}
		appList = append(appList, instanceResp)
	}
	return appList, nil
}
