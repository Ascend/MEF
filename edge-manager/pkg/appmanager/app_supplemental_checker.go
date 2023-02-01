// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager service
package appmanager

import (
	"errors"
	"fmt"
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

type deployParaChecker struct {
	req *DeployAppReq
}

func (c *deployParaChecker) Check() error {
	if len(c.req.NodeGroupIds) == 0 {
		return errors.New("deploy group ids is empty")
	}
	var checkItems = []func() error{
		c.checkDuplicate,
		c.checkIsDeployed,
		c.checkGroupIdExist,
	}
	for _, checkItem := range checkItems {
		if err := checkItem(); err != nil {
			return err
		}
	}
	return nil
}

func (c *deployParaChecker) checkDuplicate() error {
	nodeGroupIDs := make(map[uint64]struct{})
	for _, id := range c.req.NodeGroupIds {
		if _, ok := nodeGroupIDs[id]; ok {
			return errors.New("duplicated deploy group ids")
		}
		nodeGroupIDs[id] = struct{}{}
	}
	return nil
}

func (c *deployParaChecker) checkIsDeployed() error {
	deployedNodeGroupInfos, err := AppRepositoryInstance().getNodeGroupInfosByAppID(c.req.AppID)
	if err != nil {
		return errors.New("get deployed node group info, db error")
	}
	deployedGroupMap := make(map[uint64]string)
	for _, nodeGroupInfo := range deployedNodeGroupInfos {
		deployedGroupMap[nodeGroupInfo.NodeGroupID] = nodeGroupInfo.NodeGroupName
	}
	for _, id := range c.req.NodeGroupIds {
		name, ok := deployedGroupMap[id]
		if ok {
			return fmt.Errorf("check group id %v error, group name %v already deployed", id, name)
		}
	}
	return nil
}

func (c *deployParaChecker) checkGroupIdExist() error {
	nodeGroupInfos, err := getNodeGroupInfos(c.req.NodeGroupIds)
	if err != nil {
		return fmt.Errorf("get legal node group info error: %v", err)
	}
	legalGroupMap := make(map[uint64]struct{})
	for _, nodeGroupInfo := range nodeGroupInfos {
		legalGroupMap[nodeGroupInfo.NodeGroupID] = struct{}{}
	}
	for _, id := range c.req.NodeGroupIds {
		_, ok := legalGroupMap[id]
		if !ok {
			return fmt.Errorf("check group id %v error, no such group exist", id)
		}
	}
	return nil
}

type undeployParaParser struct {
	req *UndeployAppReq
}

func (c *undeployParaParser) Parse() error {
	c.removeDuplicate()
	var parseItems = []func() error{
		c.parseLegalGroupIds,
	}
	for _, parseItem := range parseItems {
		if err := parseItem(); err != nil {
			return err
		}
	}
	return nil
}

func (c *undeployParaParser) parseLegalGroupIds() error {
	nodeGroupInfos, err := AppRepositoryInstance().getNodeGroupInfosByAppID(c.req.AppID)
	if err != nil {
		return errors.New("get node group info, db error")
	}
	deployedGroup := make(map[uint64]struct{})
	for _, nodeGroupInfo := range nodeGroupInfos {
		deployedGroup[nodeGroupInfo.NodeGroupID] = struct{}{}
	}
	var nodeGroupIds []uint64
	for _, id := range c.req.NodeGroupIds {
		_, ok := deployedGroup[id]
		if !ok {
			continue
		}
		nodeGroupIds = append(nodeGroupIds, id)
	}
	c.req.NodeGroupIds = nodeGroupIds
	return nil
}

func (c *undeployParaParser) removeDuplicate() {
	nodeGroupIdMap := make(map[uint64]struct{})
	var nodeGroupIds []uint64
	for _, id := range c.req.NodeGroupIds {
		if _, ok := nodeGroupIdMap[id]; ok {
			continue
		}
		nodeGroupIds = append(nodeGroupIds, id)
		nodeGroupIdMap[id] = struct{}{}
	}
	c.req.NodeGroupIds = nodeGroupIds
}
