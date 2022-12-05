// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager service
package appmanager

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/nodemanager"
	"edge-manager/pkg/util"
)

// CreateApp Create application
func CreateApp(input interface{}) common.RespMsg {
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
	app, err := getAppInfo(req)
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

// QueryApp app info
func QueryApp(input interface{}) common.RespMsg {
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

func getAppInfo(req CreateAppReq) (*AppInfo, error) {
	containers, err := json.Marshal(req.Containers)
	if err != nil {
		hwlog.RunLog.Error("marshal containers failed")
		return nil, err
	}
	return &AppInfo{
		AppName:     req.AppName,
		Description: req.Description,
		Containers:  string(containers),
		CreatedAt:   time.Now().Format(common.TimeFormat),
		ModifiedAt:  time.Now().Format(common.TimeFormat),
	}, nil
}

// ListAppInfo get appInfo list
func ListAppInfo(input interface{}) common.RespMsg {
	hwlog.RunLog.Infof("start list app infos")
	req, ok := input.(util.ListReq)
	if !ok {
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

	total, err := AppRepositoryInstance().countListAppsInfo(req.Name)
	if err != nil {
		hwlog.RunLog.Error("count apps Infos list failed")
		return common.RespMsg{Status: "", Msg: "count apps Infos list failed", Data: nil}
	}
	apps.Total = total
	hwlog.RunLog.Info("list apps Infos success")
	return common.RespMsg{Status: common.Success, Msg: "list apps Infos success", Data: apps}
}

// DeployApp deploy application on node group
func DeployApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start deploy app")
	var req DeployAppReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	checker := appDeployParaChecker{req: &req}
	if err := checker.Check(); err != nil {
		hwlog.RunLog.Errorf("app deploy para check failed: %s", err.Error())
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	appInfo, err := AppRepositoryInstance().getAppInfoById(req.AppId)
	if err != nil {
		hwlog.RunLog.Error("get app information failed")
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	for _, nodeGroup := range req.NodeGroupInfo {
		daemonSet, err := InitDaemonSet(appInfo, nodeGroup)
		if err != nil {
			hwlog.RunLog.Errorf("app daemonSet init failed: %s", err.Error())
			return common.RespMsg{Status: "", Msg: "app daemonSet init failed", Data: nil}
		}
		daemonSet, err = kubeclient.GetKubeClient().CreateDaemonSet(daemonSet)
		if err != nil {
			hwlog.RunLog.Errorf("app daemonSet create failed: %s", err.Error())
			return common.RespMsg{Status: "", Msg: "app daemonSet create failed", Data: nil}
		}
		hwlog.RunLog.Infof("%s daemonSet create on node group %s", appInfo.AppName, nodeGroup.NodeGroupName)
	}

	hwlog.RunLog.Info("all app daemonSets create success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func updateNodeGroupDaemonSet(appInfo *AppInfo, nodeGroups []NodeGroupInfo) error {
	for _, nodeGroup := range nodeGroups {
		daemonSet, err := InitDaemonSet(appInfo, nodeGroup)
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

// UpdateApp update application
func UpdateApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start update app")
	var req CreateAppReq
	var err error
	if err = common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	checker := appCreatParaChecker{req: &req}
	if err := checker.Check(); err != nil {
		hwlog.RunLog.Errorf("app update para check failed: %s", err.Error())
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	appInfo, err := getAppInfo(req)
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
		hwlog.RunLog.Error("update node group daemon set failed ")
		return common.RespMsg{Status: "", Msg: "update node group daemon set failed", Data: nil}
	}

	hwlog.RunLog.Info("app daemonSet update success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

// DeleteApp delete application by appName
func DeleteApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start delete app")
	var req DeleteAppReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	for _, appId := range req.AppIdList {
		if err := AppRepositoryInstance().deleteAppById(appId); err != nil {
			hwlog.RunLog.Error("app db delete failed")
			return common.RespMsg{Status: "", Msg: "app db delete failed", Data: nil}
		}
	}

	hwlog.RunLog.Info("app db delete success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

// ListAppInstances get deployed apps' list
func ListAppInstances(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start list app instances")
	var appInstanceResp []AppInstanceResp
	var appId uint64
	appId, ok := input.(uint64)
	if !ok {
		hwlog.RunLog.Error("list app instances failed, param type is not integer")
		return common.RespMsg{Status: "", Msg: "param type is not integer", Data: nil}
	}
	deployedApps, err := AppRepositoryInstance().listAppInstances(appId)
	if err != nil {
		hwlog.RunLog.Error("list app instances db failed")
		return common.RespMsg{Status: "", Msg: "list app instances db failed", Data: nil}
	}
	nodeStatusService := nodemanager.NodeStatusServiceInstance()
	for _, instance := range deployedApps {
		nodeName := instance.NodeName
		nodeStatus := nodeStatusService.GetNodeStatus(nodeName)
		resp := AppInstanceResp{
			AppName:       instance.AppName,
			NodeGroupName: instance.NodeGroupName,
			NodeName:      nodeName,
			NodeStatus:    nodeStatus,
			AppStatus:     instance.Status,
		}
		appInstanceResp = append(appInstanceResp, resp)
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: appInstanceResp}
}

// InitDaemonSet init daemonSet
func InitDaemonSet(appInfo *AppInfo, nodeInfo NodeGroupInfo) (*appv1.DaemonSet, error) {
	containers, err := getContainers(appInfo)
	if err != nil {
		hwlog.RunLog.Error("app daemonSet get containers failed")
		return nil, err
	}
	tmpSpec := v1.PodSpec{}
	tmpSpec.Containers = containers
	tmpSpec.NodeSelector = map[string]string{
		common.NodeGroupLabelPrefix + strconv.FormatInt(nodeInfo.NodeGroupID, DecimalScale): nodeInfo.NodeGroupName,
	}
	template := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				common.AppManagerName: AppLabel,
				AppName:               appInfo.AppName,
				AppId:                 strconv.FormatInt(int64(appInfo.ID), DecimalScale),
			},
		},
		Spec: tmpSpec,
	}
	return &appv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: appInfo.AppName,
		},
		Spec: appv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					common.AppManagerName: AppLabel,
					AppName:               appInfo.AppName,
					AppId:                 strconv.FormatInt(int64(appInfo.ID), DecimalScale),
				},
			},
			Template: template,
		},
	}, nil
}

func getContainers(appContainer *AppInfo) ([]v1.Container, error) {
	var containerInfos []Container
	if err := json.Unmarshal([]byte(appContainer.Containers), &containerInfos); err != nil {
		hwlog.RunLog.Error("app containers unmarshal failed")
		return nil, err
	}
	var containers []v1.Container
	for _, containerInfo := range containerInfos {
		resources, err := getResources(containerInfo)
		if err != nil {
			hwlog.RunLog.Error("app daemonSet get resource failed")
			return nil, err
		}

		containers = append(containers, v1.Container{
			Name:            containerInfo.Name,
			Image:           containerInfo.Image + ":" + containerInfo.ImageVersion,
			ImagePullPolicy: v1.PullIfNotPresent,
			Command:         containerInfo.Command,
			Args:            containerInfo.Args,
			Env:             getEnv(containerInfo.Env),
			Ports:           getPorts(containerInfo.Ports),
			Resources:       resources,
		})
	}
	return containers, nil
}

func getPorts(containerPorts []ContainerPort) []v1.ContainerPort {
	var ports []v1.ContainerPort
	for _, port := range containerPorts {
		ports = append(ports, v1.ContainerPort{
			Name:          port.Name,
			HostPort:      port.HostPort,
			ContainerPort: port.ContainerPort,
			Protocol:      v1.Protocol(port.Proto),
			HostIP:        port.HostIp,
		})
	}
	return ports
}

func getEnv(envInfo []EnvVar) []v1.EnvVar {
	var envs []v1.EnvVar
	for _, env := range envInfo {
		envs = append(envs, v1.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		})
	}
	return envs
}

func getResources(appContainer Container) (v1.ResourceRequirements, error) {
	var Requests map[v1.ResourceName]resource.Quantity
	var limits map[v1.ResourceName]resource.Quantity
	var device v1.ResourceName

	cpuRequest, err := resource.ParseQuantity(appContainer.CpuRequest)
	if err != nil {
		hwlog.RunLog.Error("parse cpu request failed")
		return v1.ResourceRequirements{}, err
	}
	memRequest, err := resource.ParseQuantity(appContainer.MemRequest)
	if err != nil {
		hwlog.RunLog.Error("parse memory request failed")
		return v1.ResourceRequirements{}, err
	}
	if appContainer.Npu != "" {
		device = DeviceType
		deviceValue, err := resource.ParseQuantity(appContainer.Npu)
		if err != nil {
			hwlog.RunLog.Error("parse npu resource failed")
			return v1.ResourceRequirements{}, err
		}
		Requests = map[v1.ResourceName]resource.Quantity{
			v1.ResourceCPU: cpuRequest, v1.ResourceMemory: memRequest, device: deviceValue}
		limits = map[v1.ResourceName]resource.Quantity{
			v1.ResourceCPU: cpuRequest, v1.ResourceMemory: memRequest, device: deviceValue}
	} else {
		Requests = map[v1.ResourceName]resource.Quantity{v1.ResourceCPU: cpuRequest, v1.ResourceMemory: memRequest}
		limits = map[v1.ResourceName]resource.Quantity{v1.ResourceCPU: cpuRequest, v1.ResourceMemory: memRequest}
	}
	limits, err = getLimits(appContainer.CpuLimit, appContainer.MemLimit, limits)
	if err != nil {
		hwlog.RunLog.Error("get limits resource failed")
		return v1.ResourceRequirements{}, err
	}
	return v1.ResourceRequirements{
		Limits:   limits,
		Requests: Requests,
	}, nil
}

func getLimits(cpuLimit string, memLimit string, limitMap map[v1.ResourceName]resource.Quantity) (
	map[v1.ResourceName]resource.Quantity, error) {
	if limitMap == nil {
		return nil, fmt.Errorf("limit map is nil")
	}
	if cpuLimit != "" {
		res, err := resource.ParseQuantity(cpuLimit)
		if err != nil {
			hwlog.RunLog.Error("parse cpu limits failed")
			return limitMap, err
		}
		limitMap[v1.ResourceCPU] = res
	}
	if memLimit != "" {
		res, err := resource.ParseQuantity(memLimit)
		if err != nil {
			hwlog.RunLog.Error("parse memory limits failed")
			return limitMap, err
		}
		limitMap[v1.ResourceMemory] = res
	}
	return limitMap, nil
}
