// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager service
package appmanager

import (
	"errors"
)

// NewAppSupplementalChecker [method] for getting app backend checker
func NewAppSupplementalChecker(req CreateAppReq) *appParaChecker {
	return &appParaChecker{req: &req}
}

// NewTemplateSupplementalChecker [method] for getting template backend checker
func NewTemplateSupplementalChecker(req CreateTemplateReq) *templateParaChecker {
	return &templateParaChecker{req: &req}
}

type containerParaChecker struct {
	container *Container
}

func (c *containerParaChecker) checkContainerCpuQuantityValid() error {
	if c.container.CpuLimit != nil &&
		*c.container.CpuLimit < c.container.CpuRequest {
		return errors.New("container cpu request is illegally greater than limit")
	}
	return nil
}

func (c *containerParaChecker) checkContainerMemoryQuantityValid() error {
	if c.container.MemLimit != nil &&
		*c.container.MemLimit < c.container.MemRequest {
		return errors.New("container memory request is illegally greater than limit")
	}
	return nil
}

func (c *containerParaChecker) checkContainerEnvValid() error {
	var envNames = map[string]struct{}{}
	for idx := range c.container.Env {
		if _, ok := envNames[c.container.Env[idx].Name]; ok {
			return errors.New("container env value name is not unique")
		}
		envNames[c.container.Env[idx].Name] = struct{}{}
	}
	return nil
}

func (c *containerParaChecker) check() error {
	var checkItems = []func() error{
		c.checkContainerCpuQuantityValid,
		c.checkContainerMemoryQuantityValid,
		c.checkContainerEnvValid,
	}
	for _, checkItem := range checkItems {
		if err := checkItem(); err != nil {
			return err
		}
	}
	return nil
}

type appParaChecker struct {
	req *CreateAppReq
}

func (c *appParaChecker) checkAppContainersValid() error {
	containerName := make(map[string]struct{})
	for idx := range c.req.Containers {
		if _, ok := containerName[c.req.Containers[idx].Name]; ok {
			return errors.New("check containers par failed: duplicated name")
		}
		containerName[c.req.Containers[idx].Name] = struct{}{}
		var checker = containerParaChecker{container: &c.req.Containers[idx]}
		if err := checker.check(); err != nil {
			return err
		}
	}
	return nil
}

// Check [method] for starting checker
func (c *appParaChecker) Check() error {
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

type templateParaChecker struct {
	req *CreateTemplateReq
	appParaChecker
}

func (c *templateParaChecker) checkTemplateContainersValid() error {
	c.appParaChecker.req = &CreateAppReq{
		Containers: c.req.Containers,
	}
	if err := c.appParaChecker.Check(); err != nil {
		return err
	}
	return nil
}

// Check [method] for starting checker
func (c *templateParaChecker) Check() error {
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
