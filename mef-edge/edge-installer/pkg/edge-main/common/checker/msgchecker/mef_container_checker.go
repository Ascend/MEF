// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package msgchecker

import (
	"fmt"

	"edge-installer/pkg/edge-main/common/checker/msgchecker/types"
)

type mefContainerChecker struct {
	containerChecker
}

func (c mefContainerChecker) check(pod *types.Pod) error {
	var checkFuncs []func(c *types.Container) error

	checkFuncs = []func(c *types.Container) error{
		c.checkPortMappingPara,
		c.checkContainerEnv,
		c.checkContainerResource,
		c.checkContainerVolumeMount,
		c.checkContainerProbePara,
		c.checkContainerSecurityContext,
		c.checkResourceLimits,
	}

	var containerNames = map[string]struct{}{}
	for _, container := range pod.Spec.Containers {
		for _, function := range checkFuncs {
			if err := function(&container); err != nil {
				return err
			}
		}
		if _, ok := containerNames[container.Name]; ok {
			return fmt.Errorf("container name [%s] is not unique", container.Name)
		}
		containerNames[container.Name] = struct{}{}
	}
	return nil
}

func (c mefContainerChecker) checkContainerProbePara(container *types.Container) error {
	if container.LivenessProbe != nil || container.ReadinessProbe != nil {
		return fmt.Errorf("cur config not support prob")
	}

	return nil
}

func (c mefContainerChecker) checkPortMappingPara(container *types.Container) error {
	for _, port := range container.Ports {
		if err := checkHostIP(&port); err != nil {
			return err
		}
	}
	return nil
}

func (c mefContainerChecker) checkResourceLimits(container *types.Container) error {
	// FD assumes that the default values of Limit fields are Request fields.
	// MEF-Center ensures limits fields are explicitly set.
	for name := range container.Resources.Req {
		if _, ok := container.Resources.Lim[name]; !ok {
			return fmt.Errorf("config %s is unlimited", name)
		}
	}
	return nil
}
