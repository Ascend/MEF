// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager database table
package appmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"huawei.com/mindx/common/hwlog"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"edge-manager/pkg/kubeclient"
	"huawei.com/mindxedge/base/common"
)

func formatDaemonSetName(appName string, nodeGroupId uint64) string {
	return fmt.Sprintf("%s-%s", appName, strconv.FormatUint(nodeGroupId, DecimalScale))
}

func getPodSpec(containersStr string, nodeGroupId uint64) (*v1.PodSpec, error) {
	var containerInfos []Container
	if err := json.Unmarshal([]byte(containersStr), &containerInfos); err != nil {
		return nil, errors.New("app containers unmarshal failed")
	}

	containers, err := getContainers(containerInfos)
	if err != nil {
		return nil, errors.New("app daemonSet get containers failed")
	}
	Volumes := getVolumes(containerInfos)
	reference := v1.LocalObjectReference{Name: kubeclient.DefaultImagePullSecretKey}

	tmpSpec := v1.PodSpec{}
	tmpSpec.AutomountServiceAccountToken = new(bool)
	tmpSpec.Containers = containers
	tmpSpec.ImagePullSecrets = []v1.LocalObjectReference{reference}
	tmpSpec.NodeSelector = map[string]string{
		common.NodeGroupLabelPrefix + strconv.FormatUint(nodeGroupId, DecimalScale): "",
	}
	tmpSpec.Volumes = Volumes
	return &tmpSpec, nil
}

// initDaemonSet init daemonSet
func initDaemonSet(appInfo *AppInfo, nodeGroupId uint64) (*appv1.DaemonSet, error) {
	var tmpSpec *v1.PodSpec
	var err error
	if tmpSpec, err = getPodSpec(appInfo.Containers, nodeGroupId); err != nil {
		hwlog.RunLog.Errorf("get pod spec failed: %s", err)
		return nil, err
	}
	template := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				common.AppManagerName: AppLabel,
				AppName:               appInfo.AppName,
				AppId:                 strconv.FormatInt(int64(appInfo.ID), DecimalScale),
			},
		},
		Spec: *tmpSpec,
	}
	const maxPercentage = "100%"
	maxUnavailablePod := intstr.FromString(maxPercentage)
	return &appv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: formatDaemonSetName(appInfo.AppName, nodeGroupId),
			Labels: map[string]string{
				common.AppManagerName: AppLabel,
			},
		},
		Spec: appv1.DaemonSetSpec{
			UpdateStrategy: appv1.DaemonSetUpdateStrategy{
				Type: appv1.RollingUpdateDaemonSetStrategyType,
				RollingUpdate: &appv1.RollingUpdateDaemonSet{
					MaxUnavailable: &maxUnavailablePod,
				},
			},
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

func getVolumes(containerInfos []Container) []v1.Volume {
	// containers内挂载卷名称相同，也只算一个
	var nameMap = make(map[string]struct{}) // key: host path name / configmap name
	var infosVolNameUnique []Container
	for _, containerInfo := range containerInfos {
		for _, hostPathVolume := range containerInfo.HostPathVolumes {
			if _, ok := nameMap[hostPathVolume.Name]; ok {
				continue
			}
			nameMap[hostPathVolume.Name] = struct{}{}
			infosVolNameUnique = append(infosVolNameUnique, containerInfo)
		}

		for _, hostPathVolume := range containerInfo.HostPathVolumes {
			if _, ok := nameMap[hostPathVolume.Name]; ok {
				continue
			}
			nameMap[hostPathVolume.Name] = struct{}{}
			infosVolNameUnique = append(infosVolNameUnique, containerInfo)
		}
	}

	// get volumes from container infos
	var volumes []v1.Volume
	for _, containerInfo := range infosVolNameUnique {
		volumes = append(volumes, getVolumesFromContainerInfo(containerInfo)...)
	}

	return volumes
}

func getVolumesFromContainerInfo(containerInfo Container) []v1.Volume {
	var Volumes []v1.Volume

	// host path volume
	for _, hostPathVolume := range containerInfo.HostPathVolumes {
		hostPathSource := &v1.HostPathVolumeSource{
			Path: hostPathVolume.HostPath,
		}
		Volumes = append(Volumes, v1.Volume{
			Name:         hostPathVolume.Name,
			VolumeSource: v1.VolumeSource{HostPath: hostPathSource},
		})
	}

	// configmap volume
	for _, configmapVolume := range containerInfo.ConfigmapVolumes {
		var localObjectRef = v1.LocalObjectReference{
			Name: configmapVolume.ConfigmapName,
		}

		var cmVolumeSource = &v1.ConfigMapVolumeSource{
			LocalObjectReference: localObjectRef,
		}

		Volumes = append(Volumes, v1.Volume{
			Name: configmapVolume.Name,
			VolumeSource: v1.VolumeSource{
				ConfigMap: cmVolumeSource,
			},
		})
	}

	return Volumes
}

func getContainers(containerInfos []Container) ([]v1.Container, error) {
	var containers []v1.Container
	for _, containerInfo := range containerInfos {
		resources, err := getResources(containerInfo)
		if err != nil {
			hwlog.RunLog.Error("app daemonSet get resource failed")
			return nil, err
		}

		volumes := getVolumeMounts(containerInfo)
		runAsNonRoot := true
		readOnlyRootFilesystem := true
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
				RunAsUser:              RunAsUser,
				RunAsGroup:             RunAsGroup,
				RunAsNonRoot:           &runAsNonRoot,
				ReadOnlyRootFilesystem: &readOnlyRootFilesystem,
				Capabilities: &v1.Capabilities{
					Add:  nil,
					Drop: []v1.Capability{"All"},
				},
			},
		})
	}
	return containers, nil
}

func getVolumeMounts(container Container) []v1.VolumeMount {
	var mounts []v1.VolumeMount

	// host path volume mount
	for _, hostPathVolume := range container.HostPathVolumes {
		mounts = append(mounts, v1.VolumeMount{
			Name:      hostPathVolume.Name,
			ReadOnly:  true,
			MountPath: hostPathVolume.MountPath,
		})
	}

	// configmap volume
	for _, volumeMount := range container.ConfigmapVolumes {
		mounts = append(mounts, v1.VolumeMount{
			Name:      volumeMount.Name,
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
