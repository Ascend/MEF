// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager service
package appmanager

import (
	"errors"
	"fmt"
	"math/big"
	"net"

	"edge-manager/pkg/util"
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
	if big.NewFloat(c.container.CpuRequest).Cmp(big.NewFloat(minCpuQuantity)) < 0 ||
		big.NewFloat(c.container.CpuRequest).Cmp(big.NewFloat(maxCpuQuantity)) > 0 {
		return errors.New("cpu request quantity not in valid value")
	}

	if big.NewFloat(c.container.CpuLimit).Cmp(big.NewFloat(0)) == 0 {
		return nil
	}

	if big.NewFloat(c.container.CpuLimit).Cmp(big.NewFloat(minCpuQuantity)) < 0 ||
		big.NewFloat(c.container.CpuLimit).Cmp(big.NewFloat(maxCpuQuantity)) > 0 {
		return errors.New("cpu limit quantity not in valid value")
	}

	return nil
}

func (c *containerParaChecker) checkContainerMemoryQuantityValid() error {
	if c.container.MemRequest < minMemoryQuantity || c.container.MemRequest > maxMemoryQuantity {
		return errors.New("cpu request quantity not in valid value")
	}

	if c.container.MemLimit == 0 {
		return nil
	}
	if c.container.MemLimit < minMemoryQuantity || c.container.MemLimit > maxMemoryQuantity {
		return errors.New("cpu limit quantity not in valid value")
	}

	return nil
}

func (c *containerParaChecker) checkContainerNpuQuantityValid() error {
	if big.NewFloat(c.container.CpuLimit).Cmp(big.NewFloat(0)) == 0 {
		return nil
	}

	if big.NewFloat(c.container.CpuRequest).Cmp(big.NewFloat(minNpuQuantity)) < 0 ||
		big.NewFloat(c.container.CpuRequest).Cmp(big.NewFloat(maxNpuQuantity)) > 0 {
		return errors.New("cpu limit quantity not in valid value")
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
	if c.port.HostIP == "" || c.port.HostIP == "0.0.0.0" || c.port.HostIP == "255.255.255.255" {
		return fmt.Errorf("container port host ip invalid")
	}

	ip := net.ParseIP(c.port.HostIP)
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
	if c.container.UserID < minUserId || c.container.UserID > maxUserId {
		return fmt.Errorf("container user id valid")
	}

	return nil
}

func (c *containerParaChecker) checkGroupIdValid() error {
	if c.container.UserID < minGroupId || c.container.UserID > maxGroupId {
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
