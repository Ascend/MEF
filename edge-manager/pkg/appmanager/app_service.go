// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager service
package appmanager

import (
	"encoding/json"
	"strings"
	"time"

	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/util"

	"gorm.io/gorm"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
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
	app := getAppInfo(req)
	containers, err := getAppContainer(req, app.CreatedAt, app.ModifiedAt)
	if err != nil {
		hwlog.RunLog.Error("get container list failed ")
		return common.RespMsg{Status: "", Msg: "get container list failed", Data: nil}
	}
	if err = AppRepositoryInstance().createApp(app, containers); err != nil {
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

func getAppInfo(req util.CreateAppReq) *AppInfo {
	return &AppInfo{
		AppName:     req.AppName,
		Description: req.Description,
		CreatedAt:   time.Now().Format(common.TimeFormat),
		ModifiedAt:  time.Now().Format(common.TimeFormat),
	}
}

func getAppContainer(req util.CreateAppReq, createdAt, modifiedAt string) ([]*AppContainer, error) {
	var containers []*AppContainer
	for _, container := range req.Containers {
		envs, err := marshalEnv(container.Env)
		if err != nil {
			return nil, err
		}
		containers = append(containers, &AppContainer{
			AppName:       req.AppName,
			CreatedAt:     createdAt,
			ModifiedAt:    modifiedAt,
			ContainerName: container.ContainerName,
			CpuRequest:    container.CpuRequest,
			CpuLimit:      container.CpuLimit,
			MemoryRequest: container.MemRequest,
			MemoryLimit:   container.MemLimit,
			Npu:           container.Npu,
			ImageName:     container.ImageName,
			ImageVersion:  container.ImageVersion,
			ContainerPort: container.ContainerPort,
			Command:       parseCommand(container.Command),
			Env:           envs,
			UserID:        container.UserId,
			GroupID:       container.GroupId,
			HostIp:        container.HostIp,
			HostPort:      container.HostPort,
		})
	}
	return containers, nil
}

func marshalEnv(envs []util.EnvReq) (string, error) {
	res := ""
	for _, env := range envs {
		tmp, err := json.Marshal(env)
		if err != nil {
			return "", err
		}
		res += string(tmp) + ";"
	}
	return res[:len(res)-1], nil
}

func unMarshalEnv(s string) ([]util.EnvReq, error) {
	var envReqs []util.EnvReq
	envStrs := strings.Split(s, ";")
	for _, envStr := range envStrs {
		var envReq util.EnvReq
		err := json.Unmarshal([]byte(envStr), &envReq)
		if err != nil {
			return nil, err
		}
		envReqs = append(envReqs, envReq)
	}
	return envReqs, nil
}

func parseCommand(reqCommand []string) string {
	res := ""
	for _, str := range reqCommand {
		res += str + ";"
	}
	return res[:len(res)-1]
}

// ListAppInfo get appInfo list
func ListAppInfo(input interface{}) common.RespMsg {
	hwlog.RunLog.Infof("start list deployed apps")
	var req util.ListReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
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
	hwlog.RunLog.Error("list deployed apps failed")
	return common.RespMsg{Status: "", Msg: "list deployed apps failed", Data: nil}
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

// InitDaemonSet init daemonset
func InitDaemonSet(app *AppInstanceInfo) (*appv1.DaemonSet, error) {
	container, err := getContainer(app.AppContainer)
	if err != nil {
		hwlog.RunLog.Error("app daemonset get container failed")
		return nil, err
	}
	tmpSpec := v1.PodSpec{}
	tmpSpec.Containers = container
	tmpSpec.NodeSelector = map[string]string{
		"appmanager": "test",
	}
	template := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"appmanager": "v1",
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
					"appmanager": "v1",
				},
			},
			Template: template,
		},
	}, nil
}

func getContainer(appContainer AppContainer) ([]v1.Container, error) {
	resource, err := getResources(appContainer)
	if err != nil {
		hwlog.RunLog.Error("app daemonset get resource failed")
		return nil, err
	}

	return []v1.Container{
		{
			Name:            appContainer.AppName,
			Image:           appContainer.ImageName,
			ImagePullPolicy: v1.PullIfNotPresent,
			Command:         strings.Split(appContainer.Command, ";"),
			Resources:       resource,
		},
	}, nil
}

func getResources(appContainer AppContainer) (v1.ResourceRequirements, error) {
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
	memRequest, err := resource.ParseQuantity(appContainer.MemoryRequest)
	if err != nil {
		hwlog.RunLog.Error("parse memory request failed")
		return v1.ResourceRequirements{}, err
	}
	memLimits, err := resource.ParseQuantity(appContainer.MemoryLimit)
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
