// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager service
package appmanager

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/config"
	"edge-manager/pkg/util"
)

const (
	maxVolumesNum      = 256
	maxContentValueLen = 2048

	hostPathVolumeType  = "hostPath"
	configMapVolumeType = "configMap"
)

// NewAppSupplementalChecker [method] for getting app backend checker
func NewAppSupplementalChecker(req CreateAppReq) *appParamChecker {
	return &appParamChecker{req: &req}
}

// NewTemplateSupplementalChecker [method] for getting template backend checker
func NewTemplateSupplementalChecker(req CreateTemplateReq) *templateParamChecker {
	return &templateParamChecker{req: &req}
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

func isCmExist(cmName string) error {
	if _, err := CmRepositoryInstance().queryCmByName(cmName); err != nil {
		if err == gorm.ErrRecordNotFound {
			hwlog.RunLog.Errorf("configmap [%s] does not exist", cmName)
			return fmt.Errorf("configmap [%s] does not exist", cmName)
		}

		hwlog.RunLog.Errorf("query whether the configmap [%s] exists failed, error: %v", cmName, err)
		return fmt.Errorf("query whether the configmap [%s] exists failed", cmName)
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

type templateParamChecker struct {
	req *CreateTemplateReq
	appParamChecker
}

func (c *templateParamChecker) checkTemplateContainersValid() error {
	c.appParamChecker.req = &CreateAppReq{
		Containers: c.req.Containers,
	}
	if err := c.appParamChecker.Check(); err != nil {
		return err
	}
	return nil
}

// Check [method] for template param checker
func (c *templateParamChecker) Check() error {
	var checkItems = []func() error{
		c.checkTemplateContainersValid,
	}
	for _, checkItem := range checkItems {
		if err := checkItem(); err != nil {
			return err
		}
	}
	return nil
}

// CmParamChecker cm param checker
type CmParamChecker struct {
	req *ConfigmapReq
}

// NewCmSupplementalChecker get cm param checker
func NewCmSupplementalChecker(req ConfigmapReq) *CmParamChecker {
	return &CmParamChecker{req: &req}
}

// Check cm param supplemental checker
func (cpc *CmParamChecker) Check() error {
	var checkItems = []func() error{
		cpc.checkContentKeyUnique,
		cpc.checkContentValueLen,
	}

	for _, checkItem := range checkItems {
		if err := checkItem(); err != nil {
			return err
		}
	}

	return nil
}

// configmap content key should be unique
func (cpc *CmParamChecker) checkContentKeyUnique() error {
	cmContentKeysMap := make(map[string]struct{})
	for _, content := range cpc.req.ConfigmapContent {
		if _, ok := cmContentKeysMap[content.Name]; ok {
			return errors.New("configmap content key is duplicated")
		}
		cmContentKeysMap[content.Name] = struct{}{}
	}

	return nil
}

func (cpc *CmParamChecker) checkContentValueLen() error {
	for _, content := range cpc.req.ConfigmapContent {
		if len(content.Value) > maxContentValueLen {
			return errors.New("configmap content value length is invalid")
		}
	}
	return nil
}
