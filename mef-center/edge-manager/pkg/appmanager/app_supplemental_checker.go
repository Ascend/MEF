// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package appmanager to init app manager service
package appmanager

import (
	"errors"
	"fmt"

	"edge-manager/pkg/config"
	"edge-manager/pkg/util"
)

// NewAppSupplementalChecker [method] for getting app backend checker
func NewAppSupplementalChecker(req CreateAppReq) *appParamChecker {
	return &appParamChecker{req: &req}
}

type containerParamChecker struct {
	container *Container
}

func (c *containerParamChecker) checkContainerCpuQuantityValid() error {
	if c.container.CpuLimit != nil &&
		*c.container.CpuLimit < c.container.CpuRequest {
		return errors.New("container cpu request is illegally greater than limit")
	}
	return nil
}

func (c *containerParamChecker) checkContainerMemoryQuantityValid() error {
	if c.container.MemLimit != nil &&
		*c.container.MemLimit < c.container.MemRequest {
		return errors.New("container memory request is illegally greater than limit")
	}
	return nil
}

func (c *containerParamChecker) checkContainerEnvValid() error {
	var envNames = map[string]struct{}{}
	for idx := range c.container.Env {
		if _, ok := envNames[c.container.Env[idx].Name]; ok {
			return errors.New("container env value name is not unique")
		}
		envNames[c.container.Env[idx].Name] = struct{}{}
	}
	return nil
}

func (c *containerParamChecker) checkContainerVolume() error {
	mountPaths := map[string]struct{}{}
	volumeNames := map[string]struct{}{}
	for _, hostPathVolume := range c.container.HostPathVolumes {
		if !util.InWhiteList(hostPathVolume.HostPath, config.PodConfig.HostPath) {
			return fmt.Errorf("hostpath [%s] Verification failed: not in whitelist", hostPathVolume.HostPath)
		}
		if _, ok := mountPaths[hostPathVolume.MountPath]; ok {
			return errors.New("container volume mount path is not unique")
		}
		if _, ok := volumeNames[hostPathVolume.Name]; ok {
			return errors.New("container volume mount name is not unique")
		}
		mountPaths[hostPathVolume.MountPath] = struct{}{}
		volumeNames[hostPathVolume.Name] = struct{}{}
	}

	return nil
}

func (c *containerParamChecker) check() error {
	var checkItems = []func() error{
		c.checkContainerCpuQuantityValid,
		c.checkContainerMemoryQuantityValid,
		c.checkContainerEnvValid,
		c.checkContainerVolume,
	}
	for _, checkItem := range checkItems {
		if err := checkItem(); err != nil {
			return err
		}
	}
	return nil
}

type appParamChecker struct {
	req *CreateAppReq
}

func (c *appParamChecker) checkAppContainersValid() error {
	containerName := make(map[string]struct{})
	for idx := range c.req.Containers {
		if _, ok := containerName[c.req.Containers[idx].Name]; ok {
			return errors.New("check containers par failed: duplicated name")
		}
		containerName[c.req.Containers[idx].Name] = struct{}{}
		var checker = containerParamChecker{container: &c.req.Containers[idx]}
		if err := checker.check(); err != nil {
			return err
		}
	}
	return nil
}

// Check [method] for app param checker
func (c *appParamChecker) Check() error {
	var checkItems = []func() error{
		c.checkAppContainersValid,
	}
	for _, checkItem := range checkItems {
		if err := checkItem(); err != nil {
			return err
		}
	}
	return nil
}
