// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager service
package appmanager

import (
	"encoding/json"
	"fmt"
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
	var req util.CreateAppReq
	if err := common.ParamConvert(input, &req); err != nil {
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
	hwlog.RunLog.Info("app db create success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
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
	resp.Version = appInfo.Version
	resp.AppName = appInfo.AppName
	resp.Description = appInfo.Description

	if err = json.Unmarshal([]byte(appInfo.Containers), &resp.Containers); err != nil {
		hwlog.RunLog.Error("unmarshal containers info failed")
		return common.RespMsg{Status: "", Msg: "unmarshal containers info failed", Data: nil}
	}

	hwlog.RunLog.Info("query app success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

func getAppInfo(req util.CreateAppReq) (*AppInfo, error) {
	containers, err := json.Marshal(req.Containers)
	if err != nil {
		hwlog.RunLog.Error("marshal containers failed")
		return nil, err
	}
	return &AppInfo{
		ID:          req.AppId,
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
	if err == nil {
		hwlog.RunLog.Info("list deployed apps success")
		return common.RespMsg{Status: common.Success, Msg: "", Data: apps}
	}
	if err == gorm.ErrRecordNotFound {
		hwlog.RunLog.Info("dont have any deployed apps")
		return common.RespMsg{Status: common.Success, Msg: "dont have any deployed apps", Data: nil}
	}
	hwlog.RunLog.Error("list apps failed")
	return common.RespMsg{Status: "", Msg: "list apps failed", Data: nil}
}

// DeployApp deploy application on node group
func DeployApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start deploy app")
	var req util.DeployAppReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	appInfo, err := AppRepositoryInstance().getAppInfo(req.AppId)
	if err != nil {
		hwlog.RunLog.Error("get app information failed")
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	nodeGroup, err := AppRepositoryInstance().getNodeGroupInfo(req.NodeGroupName)
	if err != nil {
		hwlog.RunLog.Error("get node group information failed")
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	daemonSet, err := InitDaemonSet(&appInfo, nodeGroup.Label)
	if err != nil {
		hwlog.RunLog.Errorf("app daemonSet init failed: %s", err.Error())
		return common.RespMsg{Status: "", Msg: "app daemonSet init failed", Data: nil}
	}
	daemonSet, err = kubeclient.GetKubeClient().CreateDaemonSet(daemonSet)
	if err != nil {
		hwlog.RunLog.Errorf("app daemonSet create failed: %s", err.Error())
		return common.RespMsg{Status: "", Msg: "app daemonSet create failed", Data: nil}
	}

	hwlog.RunLog.Info("app daemonSet create success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func updateNodeGroupDaemonSet(appInfo *AppInfo, nodeGroups []nodemanager.NodeGroup) error {
	for _, nodeGroup := range nodeGroups {
		daemonSet, err := InitDaemonSet(appInfo, nodeGroup.Label)
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
	var req util.CreateAppReq
	var err error
	if err = common.ParamConvert(input, &req); err != nil {
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
	var req util.DeleteAppReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	if err := AppRepositoryInstance().deleteApp(req.AppId); err != nil {
		hwlog.RunLog.Error("app db delete failed")
		return common.RespMsg{Status: "", Msg: "app db delete failed", Data: nil}
	}
	hwlog.RunLog.Info("app db delete success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

// InitDaemonSet init daemonSet
func InitDaemonSet(appInfo *AppInfo, nodeLabel string) (*appv1.DaemonSet, error) {
	containers, err := getContainers(appInfo)
	if err != nil {
		hwlog.RunLog.Error("app daemonSet get containers failed")
		return nil, err
	}
	tmpSpec := v1.PodSpec{}
	tmpSpec.Containers = containers
	tmpSpec.NodeSelector = map[string]string{
		AppNodeSelectorKey: AppNodeSelectorValue,
	}
	template := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				common.AppManagerName: AppLabel,
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
				},
			},
			Template: template,
		},
	}, nil
}

func getContainers(appContainer *AppInfo) ([]v1.Container, error) {
	var containerInfos []util.Container
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
			Image:           containerInfo.Image,
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

func getPorts(containerPorts []util.ContainerPort) []v1.ContainerPort {
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

func getEnv(envInfo []util.EnvVar) []v1.EnvVar {
	var envs []v1.EnvVar
	for _, env := range envInfo {
		envs = append(envs, v1.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		})
	}
	return envs
}

func getResources(appContainer util.Container) (v1.ResourceRequirements, error) {
	var Requests map[v1.ResourceName]resource.Quantity
	var limits map[v1.ResourceName]resource.Quantity

	cpuRequest, err := resource.ParseQuantity(appContainer.CpuRequest)
	if err != nil {
		hwlog.RunLog.Error("parse cpu request failed")
		return v1.ResourceRequirements{}, err
	}
	cpuLimit, err := resource.ParseQuantity(appContainer.CpuLimit)
	if err != nil {
		hwlog.RunLog.Error("parse cpu limits failed")
		return v1.ResourceRequirements{}, err
	}
	memRequest, err := resource.ParseQuantity(appContainer.MemRequest)
	if err != nil {
		hwlog.RunLog.Error("parse memory request failed")
		return v1.ResourceRequirements{}, err
	}
	memLimits, err := resource.ParseQuantity(appContainer.MemLimit)
	if err != nil {
		hwlog.RunLog.Error("parse memory limits failed")
		return v1.ResourceRequirements{}, err
	}
	Requests = map[v1.ResourceName]resource.Quantity{v1.ResourceCPU: cpuRequest, v1.ResourceMemory: memRequest}
	limits = map[v1.ResourceName]resource.Quantity{v1.ResourceCPU: cpuLimit, v1.ResourceMemory: memLimits}

	return v1.ResourceRequirements{
		Limits:   limits,
		Requests: Requests,
	}, nil
}
