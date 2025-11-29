// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package msgchecker

import (
	"errors"
	"fmt"
	"net"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/hwlog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/common/checker/msgchecker/types"
	"edge-installer/pkg/edge-main/common/configpara"
)

const (
	npuResNameReg     = "^huawei.com/Ascend[0-9a-zA-Z]{1,64}$"
	minCpuQuantity    = "0.01"
	minMemoryQuantity = "4Mi"
)

type containerChecker struct {
	operation string
}

func isPodGraceDelete(deletionTimestamp *string) bool {
	return deletionTimestamp != nil
}

func (c containerChecker) checkContainerEnv(container *types.Container) error {
	var envNames = map[string]struct{}{}
	for _, env := range container.Env {
		if _, ok := envNames[env.Name]; ok {
			return fmt.Errorf("env name [%s] is not unique", env.Name)
		}

		envNames[env.Name] = struct{}{}
	}

	return nil
}

func (c containerChecker) checkContainerResource(container *types.Container) error {
	// only operation is update will run pod need to check resource
	if c.operation != constants.OptUpdate {
		return nil
	}
	var containerResCheckFuncs = []func(c *types.Container) error{
		checkContainerCpuResource,
		checkContainerMemoryResource,
		checkContainerNpuResources,
	}

	for _, function := range containerResCheckFuncs {
		if err := function(container); err != nil {
			hwlog.RunLog.Errorf("check container resource item failed, %s", err.Error())
			return err
		}
	}
	return nil
}

func checkContainerCpuResource(c *types.Container) error {
	if err := isResourceRequestsGreatThanLimits(c, v1.ResourceCPU); err != nil {
		return err
	}
	return nil
}

func isNpu(resName string) bool {
	if checker.GetRegChecker("", npuResNameReg, true).Check(resName).Result {
		return true
	}
	return false
}

func checkContainerNpuResources(container *types.Container) error {
	for resName := range container.Resources.Req {
		if !isNpu(resName.String()) {
			continue
		}
		if err := isResourceRequestsGreatThanLimits(container, resName); err != nil {
			return err
		}
	}
	return nil
}

func checkContainerMemoryResource(container *types.Container) error {
	if err := isResourceRequestsGreatThanLimits(container, v1.ResourceMemory); err != nil {
		return err
	}
	return nil
}

func isResourceRequestsGreatThanLimits(container *types.Container, resName v1.ResourceName) error {
	var minQuantityRes = map[v1.ResourceName]resource.Quantity{
		v1.ResourceCPU:    resource.MustParse(minCpuQuantity),
		v1.ResourceMemory: resource.MustParse(minMemoryQuantity),
	}

	request, exists := container.Resources.Req[resName]
	if !exists {
		return fmt.Errorf("%s config is not exists", resName)
	}

	if request.Cmp(resource.MustParse("0")) < 0 {
		return fmt.Errorf("%s config can't be negative", resName)
	}

	if minQuantity, ok := minQuantityRes[resName]; ok && request.Cmp(minQuantity) < 0 {
		return fmt.Errorf("%s config is not match the minimal quantity", resName)
	}

	limit, exists := container.Resources.Lim[resName]
	if !exists {
		return nil
	}

	if request.Cmp(limit) > 0 {
		return fmt.Errorf("resource [%s] request [%s] great than limit [%s]",
			resName, request.String(), limit.String())
	}
	return nil
}

func checkContainerCapability(s *types.SecurityContext) error {
	if s.Capabilities == nil || len(s.Capabilities.Add) == 0 {
		return nil
	}

	if configpara.GetPodConfig().Capability == false {
		return errors.New("check container Capability failed, cur config not support")
	}

	return nil
}

func checkContainerPrivileged(s *types.SecurityContext) error {
	if configpara.GetPodConfig().Privileged == true {
		return nil
	}

	if s.Privileged != nil && *s.Privileged == true {
		return errors.New("check container Privileged failed, cur config not support")
	}
	return nil
}

func checkContainerRunningUserPara(s *types.SecurityContext) error {
	if s.RunAsUser == nil {
		return nil
	}
	if configpara.GetPodConfig().RunAsRoot == true {
		return nil
	}
	if *s.RunAsUser == 0 {
		return errors.New("check container run as user failed, cur config not support")
	}
	return nil
}

func checkContainerRunningGroupPara(s *types.SecurityContext) error {
	if s.RunAsGroup == nil {
		return nil
	}

	if configpara.GetPodConfig().RunAsRoot == true {
		return nil
	}
	if *s.RunAsGroup == 0 {
		return errors.New("check container run as group failed, cur config not support")
	}
	return nil
}

func checkReadOnlyRootFilesystem(s *types.SecurityContext) error {
	if configpara.GetPodConfig().AllowReadWriteRootFs == true {
		return nil
	}
	if s.ReadOnlyRootFilesystem != nil && *s.ReadOnlyRootFilesystem != true {
		return errors.New("check container ReadOnlyRootFilesystem failed, not support")
	}
	return nil
}

func checkAllowPrivilegeEscalation(s *types.SecurityContext) error {
	if configpara.GetPodConfig().AllowPrivilegeEscalation == true {
		return nil
	}

	if s.AllowPrivilegeEscalation != nil && *s.AllowPrivilegeEscalation != false {
		return errors.New("check container AllowPrivilegeEscalation failed, cur config not support")
	}
	return nil
}

func checkContainerSeccomp(s *types.SecurityContext) error {
	if configpara.GetPodConfig().UseSeccomp == true {
		return nil
	}

	defaultSeccompProfile := types.SeccompProfile{Type: string(v1.SeccompProfileTypeRuntimeDefault)}

	if s.SeccompProfile != nil && *s.SeccompProfile != defaultSeccompProfile {
		return errors.New("check container seccompProfile failed, cur config not support")
	}
	return nil
}

func (c containerChecker) checkContainerSecurityContext(container *types.Container) error {
	if container.SecContext == nil {
		hwlog.RunLog.Info("container SecurityContext is nil, cur config no need to check para")
		return nil
	}

	var containerSecurityCheckFuncs = []func(SecurityContext *types.SecurityContext) error{
		checkContainerCapability,
		checkContainerPrivileged,
		checkContainerRunningUserPara,
		checkContainerRunningGroupPara,
		checkReadOnlyRootFilesystem,
		checkAllowPrivilegeEscalation,
		checkContainerSeccomp,
	}

	for index, function := range containerSecurityCheckFuncs {
		if err := function(container.SecContext); err != nil {
			hwlog.RunLog.Errorf("check container securityContext %d item failed", index)
			return err
		}
	}

	return nil
}

func (c containerChecker) checkPortMappingPara(container *types.Container) error {
	var checkFuncs = []func(port *types.ContainerPort) error{
		checkHostIP,
		checkPortUsed,
	}

	for _, port := range container.Ports {
		for _, check := range checkFuncs {
			if err := check(&port); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkPortUsed check if the configured ports have be used by other process.
// Because of the port occupation of the application itself, only fd message checks used ports.
func checkPortUsed(port *types.ContainerPort) error {
	usedPorts, err := util.GetUsedPorts(v1.Protocol(port.Protocol))
	if err != nil {
		return fmt.Errorf("check HostPort: %d failed, read system file failed, %s", port.HostPort, err.Error())
	}
	if len(usedPorts) > 0 {
		if usedPorts.Has(int64(port.HostPort)) {
			return fmt.Errorf("cur HostPort: %d is used, while its protocol is %s", port.HostPort, port.Protocol)
		}
	}

	return nil
}

// checkHostIP check host ip for container
func checkHostIP(port *types.ContainerPort) error {
	if checkResult := checker.GetIpV4Checker("HostIP", true).Check(*port); !checkResult.Result {
		return errors.New(checkResult.Reason)
	}
	validIP := map[string]struct{}{}
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return errors.New("get node ip addresses failed")
	}
	for _, address := range addresses {
		ipNet, ok := address.(*net.IPNet)
		if !ok {
			return errors.New("assert node ip failed")
		}
		if ipNet.IP.To4() == nil {
			continue
		}
		validIP[ipNet.IP.String()] = struct{}{}
	}

	_, ok := validIP[port.HostIP]
	if !ok {
		return errors.New("host ip is invalid to the node")
	}
	return nil
}

// checkContainerVolumeMount check volume mount for container
func (c containerChecker) checkContainerVolumeMount(container *types.Container) error {
	var mountPaths = map[string]struct{}{}
	var volumeNames = map[string]struct{}{}
	for _, vm := range container.VolumeMounts {
		if _, ok := mountPaths[vm.MountPath]; ok {
			return errors.New("container volume mount path is not unique")
		}

		if _, ok := volumeNames[vm.Name]; ok {
			return errors.New("container volume mount name is not unique")
		}

		mountPaths[vm.MountPath] = struct{}{}
		volumeNames[vm.Name] = struct{}{}
	}

	return nil
}
