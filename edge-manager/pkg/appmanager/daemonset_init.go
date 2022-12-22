// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager database table
package appmanager

import (
	"encoding/json"
	"fmt"
	"strconv"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func formatDaemonSetName(appName string, nodeGroupId int64) string {
	return fmt.Sprintf("%s-%s", appName, strconv.FormatInt(nodeGroupId, DecimalScale))
}

// initDaemonSet init daemonSet
func initDaemonSet(appInfo *AppInfo, nodeGroupId int64) (*appv1.DaemonSet, error) {
	containers, err := getContainers(appInfo)
	if err != nil {
		hwlog.RunLog.Error("app daemonSet get containers failed")
		return nil, err
	}
	tmpSpec := v1.PodSpec{}
	tmpSpec.Containers = containers
	tmpSpec.NodeSelector = map[string]string{
		common.NodeGroupLabelPrefix + strconv.FormatInt(nodeGroupId, DecimalScale): "",
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
			Name: formatDaemonSetName(appInfo.AppName, nodeGroupId),
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
		device = common.DeviceType
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
