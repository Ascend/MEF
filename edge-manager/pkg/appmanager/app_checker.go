// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager service
package appmanager

import (
	"errors"
	"fmt"
	"net"

	"edge-manager/pkg/util"

	"k8s.io/apimachinery/pkg/api/resource"
)

type appParaPattern struct {
	patterns map[string]string
}

var appPattern = appParaPattern{patterns: map[string]string{
	"appName":           "^[a-z]([a-z0-9-]{0,30}[a-z0-9]){0,1}$",
	"description":       `^[\S ]{0,512}$`,
	"containerName":     "^[a-z0-9]([a-z0-9-]{0,30}[a-z0-9]){0,1}$",
	"containerImage":    "^[a-z0-9]([a-z0-9_./-]{0,30}[a-z0-9]){0,1}$",
	"imageVersion":      "^[a-zA-Z0-9_.-]{1,32}$",
	"containerCommand":  "^[a-zA-Z0-9 _./-]{0,255}[a-zA-Z0-9]$",
	"containerArgs":     "^[a-zA-Z0-9 _./-]{0,255}[a-zA-Z0-9]$",
	"containerEnvName":  "^[a-zA-Z][a-zA-z0-9._-]{0,30}[a-zA-Z0-9]$",
	"containerEnvValue": "^[a-zA-Z0-9 _./-]{1,512}$",
	"containerPortName": "^[a-z0-9]([a-z0-9-]{0,30}[a-z0-9]){0,1}$",
	"nodeGroupName":     "^[a-z]([a-z0-9-]{0,30}[a-z0-9]){0,1}$",
	"templateName":      "^[a-z]([a-z0-9-]{0,30}[a-z0-9]){0,1}$",
},
}

func (a *appParaPattern) getPattern(key string) (string, bool) {
	pattern, ok := a.patterns[key]
	return pattern, ok
}

type containerParaChecker struct {
	container *Container
}

func (c *containerParaChecker) checkContainerNameValid() error {
	pattern, ok := appPattern.getPattern("containerName")
	if !ok {
		return fmt.Errorf("containerName regex pattern not exist")
	}

	if !util.RegexStringChecker(c.container.Name, pattern) {
		return fmt.Errorf("container name invalid")
	}
	return nil
}

func (c *containerParaChecker) checkContainerImageValid() error {
	pattern, ok := appPattern.getPattern("containerImage")
	if !ok {
		return fmt.Errorf("containerImage regex pattern not exist")
	}
	if !util.RegexStringChecker(c.container.Image, pattern) {
		return fmt.Errorf("container image invalid")
	}
	return nil
}

func (c *containerParaChecker) checkContainerImageVersionValid() error {
	pattern, ok := appPattern.getPattern("imageVersion")
	if !ok {
		return fmt.Errorf("imageVersion regex pattern not exist")
	}
	if !util.RegexStringChecker(c.container.ImageVersion, pattern) {
		return fmt.Errorf("container image version invalid")
	}

	return nil
}

func (c *containerParaChecker) checkContainerCpuQuantityValid() error {
	cpuRequest, err := resource.ParseQuantity(c.container.CpuRequest)
	if err != nil {
		return errors.New("parse cpu request failed")
	}

	minMem := resource.NewMilliQuantity(minCpuQuantity*milliQuantityUnit, resource.DecimalSI)
	maxMem := resource.NewMilliQuantity(maxCpuQuantity*milliQuantityUnit, resource.DecimalSI)

	if cpuRequest.Cmp(*minMem) < 0 || cpuRequest.Cmp(*maxMem) > 0 {
		return errors.New("cpu request quantity not in valid value")
	}

	if len(c.container.CpuLimit) == 0 {
		return nil
	}
	cpuLimit, err := resource.ParseQuantity(c.container.CpuLimit)
	if err != nil {
		return errors.New("parse cpu limit failed")
	}

	if cpuLimit.Cmp(*minMem) < 0 || cpuLimit.Cmp(*maxMem) > 0 {
		return errors.New("cpu request quantity not in valid value")
	}
	return nil
}

func (c *containerParaChecker) checkContainerMemoryQuantityValid() error {
	memRequest, err := resource.ParseQuantity(c.container.MemRequest)
	if err != nil {
		return errors.New("parse mem request failed")
	}

	minMem := resource.NewQuantity(minMemoryQuantity, resource.BinarySI)
	maxMem := resource.NewQuantity(maxMemoryQuantity, resource.BinarySI)

	if memRequest.Cmp(*minMem) < 0 || memRequest.Cmp(*maxMem) > 0 {
		return errors.New("mem request quantity not in valid value")
	}

	if len(c.container.MemLimit) == 0 {
		return nil
	}
	memLimit, err := resource.ParseQuantity(c.container.MemLimit)
	if err != nil {
		return errors.New("parse mem limit failed")
	}

	if memLimit.Cmp(*minMem) < 0 || memLimit.Cmp(*maxMem) > 0 {
		return errors.New("mem request quantity not in valid value")
	}
	return nil
}

func (c *containerParaChecker) checkContainerNpuQuantityValid() error {
	if len(c.container.Npu) == 0 {
		return nil
	}

	npuRequest, err := resource.ParseQuantity(c.container.Npu)
	if err != nil {
		return errors.New("parse npu request failed")
	}

	minMem := resource.NewMilliQuantity(minNpuQuantity*milliQuantityUnit, resource.DecimalSI)
	maxMem := resource.NewMilliQuantity(maxNpuQuantity*milliQuantityUnit, resource.DecimalSI)

	if npuRequest.Cmp(*minMem) < 0 || npuRequest.Cmp(*maxMem) > 0 {
		return errors.New("npu request quantity not in valid value")
	}

	return nil
}

func (c *containerParaChecker) checkContainerCommandValid() error {
	if len(c.container.Command) > commandMaxCount {
		return fmt.Errorf("container command count up to limt")
	}

	pattern, ok := appPattern.getPattern("containerCommand")
	if !ok {
		return fmt.Errorf("containerCommand regex pattern not exist")
	}

	for _, command := range c.container.Command {
		if !util.RegexStringChecker(command, pattern) {
			return fmt.Errorf("container command invalid")
		}
	}

	return nil
}

func (c *containerParaChecker) checkContainerArgsValid() error {
	if len(c.container.Args) > argsMaxCount {
		return fmt.Errorf("container args count up to limt")
	}

	pattern, ok := appPattern.getPattern("containerArgs")
	if !ok {
		return fmt.Errorf("containerArgs regex pattern not exist")
	}

	for _, arg := range c.container.Args {
		if !util.RegexStringChecker(arg, pattern) {
			return fmt.Errorf("container arg invalid")
		}
	}

	return nil
}

func (c *containerParaChecker) checkContainerEnvValid() error {
	if len(c.container.Env) > envMaxCount {
		return fmt.Errorf("container image env var num up to limit")
	}

	namePattern, ok := appPattern.getPattern("containerEnvName")
	if !ok {
		return fmt.Errorf("containerEnvName regex pattern not exist")
	}

	valuePattern, ok := appPattern.getPattern("containerEnvValue")
	if !ok {
		return fmt.Errorf("containerEnvValue regex pattern not exist")
	}

	var envNames = map[string]struct{}{}
	for idx := range c.container.Env {
		if !util.RegexStringChecker(c.container.Env[idx].Name, namePattern) {
			return fmt.Errorf("container env var name invalid")
		}

		if !util.RegexStringChecker(c.container.Env[idx].Value, valuePattern) {
			return fmt.Errorf("container env var value invalid")
		}

		if _, ok := envNames[c.container.Env[idx].Name]; ok {
			return fmt.Errorf("container env value name is not unique")
		}
		envNames[c.container.Env[idx].Name] = struct{}{}
	}

	return nil
}

func (c *containerParaChecker) check() error {
	var checkItems = []func() error{
		c.checkContainerNameValid,
		c.checkContainerImageValid,
		c.checkContainerImageVersionValid,
		c.checkContainerCpuQuantityValid,
		c.checkContainerMemoryQuantityValid,
		c.checkContainerNpuQuantityValid,
		c.checkContainerCommandValid,
		c.checkContainerArgsValid,
		c.checkContainerEnvValid,
		c.checkContainerPortsValid,
		c.checkUserIdValid,
		c.checkGroupIdValid,
	}
	for _, checkItem := range checkItems {
		if err := checkItem(); err != nil {
			return err
		}
	}
	return nil
}

type portParaChecker struct {
	port *ContainerPort
}

func (c *portParaChecker) checkPortName() error {
	pattern, ok := appPattern.getPattern("containerPortName")
	if !ok {
		return fmt.Errorf("containerPortName regex pattern not exist")
	}

	if !util.RegexStringChecker(c.port.Name, pattern) {
		return fmt.Errorf("container port name invalid")
	}
	return nil
}

func (c *portParaChecker) checkPortProtocol() error {
	if c.port.Proto != "TCP" && c.port.Proto != "UDP" && c.port.Proto != "SCTP" {
		return fmt.Errorf("container port protocol invalid")
	}

	return nil
}

func (c *portParaChecker) checkPortContainerPort() error {
	if c.port.ContainerPort < minContainerPort || c.port.ContainerPort > maxContainerPort {
		return fmt.Errorf("container port invalid")
	}

	return nil
}

func (c *portParaChecker) checkPortHostPort() error {
	if c.port.HostPort < minHostPort || c.port.HostPort > maxHostPort {
		return fmt.Errorf("container host port invalid")
	}
	return nil
}

func (c *portParaChecker) checkPortHostIP() error {
	if c.port.HostIp == "" || c.port.HostIp == "0.0.0.0" || c.port.HostIp == "255.255.255.255" {
		return fmt.Errorf("container port host ip invalid")
	}

	ip := net.ParseIP(c.port.HostIp)
	if ip == nil || ip.To4() == nil {
		return fmt.Errorf("container port host ip is not ipv4")
	}

	return nil
}

func (c *containerParaChecker) checkContainerPortsValid() error {
	if len(c.container.Ports) > portMapMaxCount {
		return fmt.Errorf("container ports num up to limit")
	}

	for idx := range c.container.Ports {
		var checker = portParaChecker{port: &c.container.Ports[idx]}
		var checkItems = []func() error{
			checker.checkPortName,
			checker.checkPortProtocol,
			checker.checkPortContainerPort,
			checker.checkPortHostPort,
			checker.checkPortHostIP,
		}
		for _, checkItem := range checkItems {
			if err := checkItem(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *containerParaChecker) checkUserIdValid() error {
	if c.container.UserId < minUserId || c.container.UserId > maxUserId {
		return fmt.Errorf("container user id valid")
	}

	return nil
}

func (c *containerParaChecker) checkGroupIdValid() error {
	if c.container.UserId < minGroupId || c.container.UserId > maxGroupId {
		return fmt.Errorf("container group id valid")
	}

	return nil
}

type appParaChecker struct {
	req *CreateAppReq
}

func (c *appParaChecker) checkAppNameValid() error {
	pattern, ok := appPattern.getPattern("appName")
	if !ok {
		return fmt.Errorf("appName regex pattern not exist")
	}

	if !util.RegexStringChecker(c.req.AppName, pattern) {
		return fmt.Errorf("app name invalid")
	}
	return nil
}

func (c *appParaChecker) checkAppDescriptionValid() error {
	pattern, ok := appPattern.getPattern("description")
	if !ok {
		return fmt.Errorf("description regex pattern not exist")
	}

	if !util.RegexStringChecker(c.req.Description, pattern) {
		return fmt.Errorf("app description invalid")
	}
	return nil
}

func (c *appParaChecker) checkAppContainersValid() error {
	for idx := range c.req.Containers {
		var checker = containerParaChecker{container: &c.req.Containers[idx]}
		if err := checker.check(); err != nil {
			return err
		}
	}

	return nil
}

func (c *appParaChecker) Check() error {
	var checkItems = []func() error{
		c.checkAppNameValid,
		c.checkAppDescriptionValid,
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
}

func (c *templateParaChecker) checkTemplateNameValid() error {
	pattern, ok := appPattern.getPattern("templateName")
	if !ok {
		return fmt.Errorf("templateName regex pattern not exist")
	}

	if !util.RegexStringChecker(c.req.Name, pattern) {
		return fmt.Errorf("app name invalid")
	}
	return nil
}

func (c *templateParaChecker) checkTemplateDescriptionValid() error {
	pattern, ok := appPattern.getPattern("description")
	if !ok {
		return fmt.Errorf("description regex pattern not exist")
	}

	if !util.RegexStringChecker(c.req.Description, pattern) {
		return fmt.Errorf("template description invalid")
	}
	return nil
}

func (c *templateParaChecker) checkTemplateContainersValid() error {
	for idx := range c.req.Containers {
		var checker = containerParaChecker{container: &c.req.Containers[idx]}
		if err := checker.check(); err != nil {
			return err
		}
	}

	return nil
}

func (c *templateParaChecker) Check() error {
	var checkItems = []func() error{
		c.checkTemplateNameValid,
		c.checkTemplateDescriptionValid,
		c.checkTemplateContainersValid,
	}
	for _, checkItem := range checkItems {
		if err := checkItem(); err != nil {
			return err
		}
	}
	return nil
}
