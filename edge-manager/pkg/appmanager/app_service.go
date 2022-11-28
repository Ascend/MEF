// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager service
package appmanager

import (
	"encoding/json"
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

func getAppInfo(req util.CreateAppReq) (*AppInfo, error) {
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

	appInstanceInfo, err := AppRepositoryInstance().getAppAndNodeGroupInfo(req.AppName, req.NodeGroupName)
	if err != nil {
		hwlog.RunLog.Error("get app and node group information failed")
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	daemonset, err := InitDaemonSet(appInstanceInfo)
	if err != nil {
		hwlog.RunLog.Error("app daemonset init failed")
		return common.RespMsg{Status: "", Msg: "app daemonset init failed", Data: nil}
	}
	daemonset, err = kubeclient.GetKubeClient().CreateDaemonSet(daemonset)
	if err != nil {
		hwlog.RunLog.Error("app daemonset create failed")
		return common.RespMsg{Status: "", Msg: "app daemonset create failed", Data: nil}
	}

	hwlog.RunLog.Info("app daemonset create success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

// DeleteApp delete application by appName
func DeleteApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start delete app")
	var req util.DeleteAppReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	if err := AppRepositoryInstance().deleteApp(req.AppName); err != nil {
		hwlog.RunLog.Error("app db delete failed")
		return common.RespMsg{Status: "", Msg: "app db delete failed", Data: nil}
	}
	hwlog.RunLog.Info("app db delete success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

// InitDaemonSet init daemonset
func InitDaemonSet(app *AppInstanceInfo) (*appv1.DaemonSet, error) {
	containers, err := getContainers(app.AppInfo)
	if err != nil {
		hwlog.RunLog.Error("app daemonset get containers failed")
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
			Name: app.AppInfo.AppName,
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

func getContainers(appContainer AppInfo) ([]v1.Container, error) {
	var containerInfos []util.ContainerReq
	if err := json.Unmarshal([]byte(appContainer.Containers), &containerInfos); err != nil {
		hwlog.RunLog.Error("app containers unmarshal failed")
		return nil, err
	}
	var containers []v1.Container
	for _, containerInfo := range containerInfos {
		resources, err := getResources(containerInfo)
		if err != nil {
			hwlog.RunLog.Error("app daemonset get resource failed")
			return nil, err
		}

		containers = append(containers, v1.Container{
			Name:            containerInfo.ContainerName,
			Image:           containerInfo.ImageName,
			ImagePullPolicy: v1.PullIfNotPresent,
			Command:         containerInfo.Command,
			Env:             getEnv(containerInfo.Env),
			Ports:           getPorts(containerInfo.ContainerPort),
			Resources:       resources,
		})
	}
	return containers, nil
}

func getPorts(containerPorts []util.PortTransfer) []v1.ContainerPort {
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

func getEnv(envInfo []util.EnvReq) []v1.EnvVar {
	var envs []v1.EnvVar
	for _, env := range envInfo {
		envs = append(envs, v1.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		})
	}
	return envs
}

func getResources(appContainer util.ContainerReq) (v1.ResourceRequirements, error) {
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
