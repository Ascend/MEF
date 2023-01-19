// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager database table
package appmanager

import (
	"encoding/json"
	"fmt"
	"strconv"

	"huawei.com/mindx/common/hwlog"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"huawei.com/mindxedge/base/common"
)

func formatDaemonSetName(appName string, nodeGroupId int64) string {
	return fmt.Sprintf("%s-%s", appName, strconv.FormatInt(nodeGroupId, DecimalScale))
}

// initDaemonSet init daemonSet
func initDaemonSet(appInfo *AppInfo, nodeGroupId int64) (*appv1.DaemonSet, error) {
	var containerInfos []Container
	if err := json.Unmarshal([]byte(appInfo.Containers), &containerInfos); err != nil {
		hwlog.RunLog.Error("app containers unmarshal failed")
		return nil, err
	}

	containers, err := getContainers(containerInfos)
	if err != nil {
		hwlog.RunLog.Error("app daemonSet get containers failed")
		return nil, err
	}
	cmVolumes := getCmVolumes(containerInfos)

	tmpSpec := v1.PodSpec{}
	tmpSpec.Containers = containers
	tmpSpec.NodeSelector = map[string]string{
		common.NodeGroupLabelPrefix + strconv.FormatInt(nodeGroupId, DecimalScale): "",
	}
	tmpSpec.Volumes = cmVolumes
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
			Labels: map[string]string{
				common.AppManagerName: AppLabel,
			},
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

func getCmVolumes(containerInfos []Container) []v1.Volume {
	var cmVolumes []v1.Volume
	for _, containerInfo := range containerInfos {
		cmVolumes = getCmVolumesFromContainerInfo(containerInfo)
	}
	return cmVolumes
}

func getCmVolumesFromContainerInfo(containerInfo Container) []v1.Volume {
	var cmVolumes []v1.Volume
	for _, volumeMount := range containerInfo.VolumeMounts {
		var localObjectRef = v1.LocalObjectReference{
			Name: volumeMount.ConfigmapName,
		}

		var cmVolumeSource = &v1.ConfigMapVolumeSource{
			LocalObjectReference: localObjectRef,
		}

		cmVolumes = append(cmVolumes, v1.Volume{
			Name: volumeMount.LocalVolumeName,
			VolumeSource: v1.VolumeSource{
				ConfigMap: cmVolumeSource,
			},
		})
	}
	return cmVolumes
}

func getContainers(containerInfos []Container) ([]v1.Container, error) {
	var containers []v1.Container
	for _, containerInfo := range containerInfos {
		resources, err := getResources(containerInfo)
		if err != nil {
			hwlog.RunLog.Error("app daemonSet get resource failed")
			return nil, err
		}

		volumes := getVolumeMounts(containerInfo.VolumeMounts)
		runAsNonRoot := true
		RunAsUser := containerInfo.UserID
		RunAsGroup := containerInfo.GroupID
		containers = append(containers, v1.Container{
			Name:            containerInfo.Name,
			Image:           containerInfo.Image + ":" + containerInfo.ImageVersion,
			ImagePullPolicy: v1.PullIfNotPresent,
			Command:         containerInfo.Command,
			Args:            containerInfo.Args,
			Env:             getEnv(containerInfo.Env),
			Ports:           getPorts(containerInfo.Ports),
			Resources:       resources,
			VolumeMounts:    volumes,
			SecurityContext: &v1.SecurityContext{
				RunAsUser:    &RunAsUser,
				RunAsGroup:   &RunAsGroup,
				RunAsNonRoot: &runAsNonRoot,
			},
		})
	}
	return containers, nil
}

func getVolumeMounts(volumeMounts []VolumeMount) []v1.VolumeMount {
	var mounts []v1.VolumeMount
	for _, volumeMount := range volumeMounts {
		mounts = append(mounts, v1.VolumeMount{
			Name:      volumeMount.LocalVolumeName,
			ReadOnly:  true,
			MountPath: volumeMount.MountPath,
		})
	}
	return mounts
}

func getPorts(containerPorts []ContainerPort) []v1.ContainerPort {
	var ports []v1.ContainerPort
	for _, port := range containerPorts {
		ports = append(ports, v1.ContainerPort{
			Name:          port.Name,
			HostPort:      port.HostPort,
			ContainerPort: port.ContainerPort,
			Protocol:      v1.Protocol(port.Proto),
			HostIP:        port.HostIP,
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

	cpuRequest, err := resource.ParseQuantity(fmt.Sprintf("%v", appContainer.CpuRequest))
	if err != nil {
		hwlog.RunLog.Error("parse cpu request failed")
		return v1.ResourceRequirements{}, err
	}
	memRequest, err := resource.ParseQuantity(fmt.Sprintf("%vM", appContainer.MemRequest))
	if err != nil {
		hwlog.RunLog.Error("parse memory request failed")
		return v1.ResourceRequirements{}, err
	}
	if appContainer.Npu != nil {
		device = common.DeviceType
		deviceValue, err := resource.ParseQuantity(fmt.Sprintf("%v", *appContainer.Npu))
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

func getLimits(cpuLimit *float64, memLimit *int64, limitMap map[v1.ResourceName]resource.Quantity) (
	map[v1.ResourceName]resource.Quantity, error) {
	if limitMap == nil {
		return nil, fmt.Errorf("limit map is nil")
	}
	if cpuLimit != nil {
		res, err := resource.ParseQuantity(fmt.Sprintf("%v", *cpuLimit))
		if err != nil {
			hwlog.RunLog.Error("parse cpu limits failed")
			return limitMap, err
		}
		limitMap[v1.ResourceCPU] = res
	}
	if memLimit != nil {
		res, err := resource.ParseQuantity(fmt.Sprintf("%vM", *memLimit))
		if err != nil {
			hwlog.RunLog.Error("parse memory limits failed")
			return limitMap, err
		}
		limitMap[v1.ResourceMemory] = res
	}
	return limitMap, nil
}
