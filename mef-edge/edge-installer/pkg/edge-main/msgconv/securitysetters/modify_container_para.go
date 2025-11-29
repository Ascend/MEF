// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package securitysetters
package securitysetters

import (
	"huawei.com/mindx/common/hwlog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/common"
	"edge-installer/pkg/edge-main/common/configpara"
)

func setContainerDropCapabilities(securityContext *v1.SecurityContext) {
	if configpara.GetPodConfig().UseDefaultContainerCap == true {
		return
	}

	var drop = []v1.Capability{"All"}
	hwlog.RunLog.Info("drop all capabilities")

	if securityContext.Capabilities == nil {
		securityContext.Capabilities = new(v1.Capabilities)
	}
	securityContext.Capabilities.Drop = drop
}

func setContainerReadOnlyRootFs(securityContext *v1.SecurityContext) {
	if configpara.GetPodConfig().AllowReadWriteRootFs == true {
		return
	}

	hwlog.RunLog.Info("set root fs to read only")

	readOnlyRootFilesystem := true
	securityContext.ReadOnlyRootFilesystem = &readOnlyRootFilesystem
}

func setContainerRunAsNonRoot(securityContext *v1.SecurityContext) {
	if configpara.GetPodConfig().RunAsRoot == true {
		return
	}

	runAsNonRoot := true
	securityContext.RunAsNonRoot = &runAsNonRoot
	hwlog.RunLog.Info("set run as not root to true")
}

func setContainerPrivileged(securityContext *v1.SecurityContext) {
	if configpara.GetPodConfig().Privileged == true {
		return
	}

	privileged := false
	securityContext.Privileged = &privileged
	hwlog.RunLog.Info("set privileged to false")
}

func setContainerAllowPrivilegeEscalation(securityContext *v1.SecurityContext) {
	if configpara.GetPodConfig().AllowPrivilegeEscalation == true {
		return
	}

	allowPrivilegeEscalation := false
	securityContext.AllowPrivilegeEscalation = &allowPrivilegeEscalation
	hwlog.RunLog.Info("set allowPrivilegeEscalation to false")
}

func setContainerSeccomp(securityContext *v1.SecurityContext) {
	if configpara.GetPodConfig().UseSeccomp == true {
		return
	}

	securityContext.SeccompProfile = &v1.SeccompProfile{Type: v1.SeccompProfileTypeRuntimeDefault}
	hwlog.RunLog.Info("set seccomp to runtime default")
}

func setContainerPort(container *v1.Container, containerTmp *v1.Container) {
	var ports []v1.ContainerPort
	for _, port := range container.Ports {
		ports = append(ports, v1.ContainerPort{
			HostPort:      port.HostPort,
			ContainerPort: port.ContainerPort,
			Protocol:      port.Protocol,
			HostIP:        port.HostIP})
	}
	containerTmp.Ports = ports
}

func setContainerEnv(container *v1.Container, containerTmp *v1.Container) {
	var envs []v1.EnvVar
	for _, env := range container.Env {
		envs = append(envs, v1.EnvVar{Name: env.Name, Value: env.Value})
	}
	containerTmp.Env = envs
}

func setContainerVolumeMount(container *v1.Container, containerTmp *v1.Container) {
	var volumeMounts []v1.VolumeMount
	for _, volumeMount := range container.VolumeMounts {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      volumeMount.Name,
			ReadOnly:  volumeMount.ReadOnly,
			MountPath: volumeMount.MountPath})
	}
	containerTmp.VolumeMounts = volumeMounts
}

func setContainerProbe(probe *v1.Probe) *v1.Probe {
	if probe == nil {
		return nil
	}

	probeTmp := &v1.Probe{
		InitialDelaySeconds: probe.InitialDelaySeconds,
		TimeoutSeconds:      probe.TimeoutSeconds,
		PeriodSeconds:       probe.PeriodSeconds,
		SuccessThreshold:    probe.SuccessThreshold,
		FailureThreshold:    probe.FailureThreshold}

	if probe.Exec != nil {
		probeTmp.Exec = probe.Exec
	} else if probe.HTTPGet != nil {
		probeTmp.HTTPGet = &v1.HTTPGetAction{
			Path:   probe.HTTPGet.Path,
			Port:   intstr.IntOrString{IntVal: probe.HTTPGet.Port.IntVal},
			Host:   probe.HTTPGet.Host,
			Scheme: probe.HTTPGet.Scheme}
	}

	return probeTmp
}

func setContainerLivenessProbe(container *v1.Container, containerTmp *v1.Container) {
	containerTmp.LivenessProbe = setContainerProbe(container.LivenessProbe)
}

func setContainerReadinessProbe(container *v1.Container, containerTmp *v1.Container) {
	containerTmp.ReadinessProbe = setContainerProbe(container.ReadinessProbe)
}

func setContainerSecurityContext(container *v1.Container, containerTmp *v1.Container) {
	if container.SecurityContext == nil {
		container.SecurityContext = new(v1.SecurityContext)
	}

	if container.SecurityContext.Capabilities == nil {
		container.SecurityContext.Capabilities = new(v1.Capabilities)
	}

	securityContextTmp := &v1.SecurityContext{
		Capabilities: &v1.Capabilities{Add: container.SecurityContext.Capabilities.Add},
		Privileged:   container.SecurityContext.Privileged,
		RunAsUser:    container.SecurityContext.RunAsUser,
		RunAsGroup:   container.SecurityContext.RunAsGroup}

	var securitySetFuncs = []func(securityContext *v1.SecurityContext){
		setContainerReadOnlyRootFs,
		setContainerDropCapabilities,
		setContainerRunAsNonRoot,
		setContainerPrivileged,
		setContainerAllowPrivilegeEscalation,
		setContainerSeccomp,
	}

	for _, setFunc := range securitySetFuncs {
		setFunc(securityContextTmp)
	}

	containerTmp.SecurityContext = securityContextTmp
}

func setContainerResourceLimit(container *v1.Container, containerTmp *v1.Container) {
	npuName, ok := common.LoadNpuFromDb()
	if !ok {
		hwlog.RunLog.Warnf("load npu name failed, use default resource [%s]", constants.ResourceNPU)
		npuName = constants.ResourceNPU
	}

	containerTmp.Resources.Requests = v1.ResourceList{}
	containerTmp.Resources.Limits = v1.ResourceList{}

	for resName := range container.Resources.Limits {
		if resName == constants.ResourceNPU {
			resName = v1.ResourceName(npuName)
		}
		if isResNameValid(resName) {
			containerTmp.Resources.Limits[resName] = container.Resources.Limits[resName]
		}
	}
	for resName := range container.Resources.Requests {
		if resName == constants.ResourceNPU {
			resName = v1.ResourceName(npuName)
		}
		if isResNameValid(resName) {
			containerTmp.Resources.Requests[resName] = container.Resources.Requests[resName]
		}

		if _, ok := containerTmp.Resources.Limits[resName]; !ok {
			containerTmp.Resources.Limits[resName] = container.Resources.Requests[resName]
		}
	}
}

func setContainerPara(podSpec, podSpecTmp *v1.PodSpec) {
	var containerSetFuncs = []func(container, containerTmp *v1.Container){
		setContainerResourceLimit,
		setContainerPort,
		setContainerEnv,
		setContainerVolumeMount,
		setContainerLivenessProbe,
		setContainerReadinessProbe,
		setContainerSecurityContext,
	}

	for _, container := range podSpec.Containers {
		var containerTmp = v1.Container{
			Name:                     container.Name,
			Image:                    container.Image,
			Command:                  container.Command,
			Args:                     container.Args,
			TerminationMessagePolicy: container.TerminationMessagePolicy,
			ImagePullPolicy:          container.ImagePullPolicy,
		}

		for _, setFunc := range containerSetFuncs {
			setFunc(&container, &containerTmp)
		}

		podSpecTmp.Containers = append(podSpecTmp.Containers, containerTmp)
	}
}

func isResNameValid(resName v1.ResourceName) bool {
	if resName == v1.ResourceCPU || resName == v1.ResourceMemory {
		return true
	}
	if util.IsWholeNpu(string(resName)) {
		return true
	}
	if resName == constants.ResourceNPU {
		return true
	}
	return false
}
