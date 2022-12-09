// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package types to init app manager service
package types

import (
	"fmt"
	"net"

	"edge-manager/pkg/util"
)

type AppParaPattern struct {
	patterns map[string]string
}

var appPattern = AppParaPattern{patterns: map[string]string{
	"appName":           "^[a-z]([a-z0-9-]{0,30}[a-z0-9]){0,1}$",
	"appDescription":    `^[\S ]{0,512}$`,
	"containerName":     "^[a-z0-9]([a-z0-9-]{0,30}[a-z0-9]){0,1}$",
	"containerImage":    "^[a-z0-9]([a-z0-9_./-]{0,30}[a-z0-9]){0,1}$",
	"imageVersion":      "^[a-zA-Z0-9_.-]{1,32}$",
	"containerCommand":  "^[a-zA-Z0-9 _./-]{0,31}[a-zA-Z0-9]$",
	"containerArgs":     "^[a-zA-Z0-9 _./-]{0,31}[a-zA-Z0-9]$",
	"containerEnvName":  "^[a-zA-Z][a-zA-z0-9._-]{0,30}[a-zA-Z0-9]$",
	"containerEnvValue": "^[a-zA-Z0-9 _./-]{0,512}$",
	"containerPortName": "^[a-z0-9]([a-z0-9-]{0,30}[a-z0-9]){0,1}$",
	"nodeGroupName":     "^[a-z]([a-z0-9-]{0,30}[a-z0-9]){0,1}$",
},
}

func (a *AppParaPattern) getPattern(key string) (string, bool) {
	pattern, ok := a.patterns[key]
	return pattern, ok
}

func (a *AppParam) checkAppNameValid() error {
	pattern, ok := appPattern.getPattern("appName")
	if !ok {
		return fmt.Errorf("appName regex pattern not exist")
	}

	if !util.RegexStringChecker(a.AppName, pattern) {
		return fmt.Errorf("app name invalid")
	}
	return nil
}

func (a *AppParam) checkAppDescriptionValid() error {
	pattern, ok := appPattern.getPattern("appDescription")
	if !ok {
		return fmt.Errorf("appDescription regex pattern not exist")
	}

	if !util.RegexStringChecker(a.Description, pattern) {
		return fmt.Errorf("app description invalid")
	}
	return nil
}

func (a *AppParam) checkContainerParaValid() error {
	for _, container := range a.Containers {
		var checkItems = []func() error{
			container.checkContainerNameValid,
			container.checkContainerImageValid,
			container.checkContainerImageVersionValid,
			container.checkContainerCommandValid,
			container.checkContainerArgsValid,
			container.checkContainerEnvValid,
			container.checkContainerPortsValid,
			container.checkUserIdValid,
			container.checkGroupIdValid,
		}
		for _, checkItem := range checkItems {
			if err := checkItem(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Container) checkContainerNameValid() error {
	pattern, ok := appPattern.getPattern("containerName")
	if !ok {
		return fmt.Errorf("containerName regex pattern not exist")
	}

	if !util.RegexStringChecker(c.Name, pattern) {
		return fmt.Errorf("container name invalid")
	}
	return nil
}

func (c *Container) checkContainerImageValid() error {
	pattern, ok := appPattern.getPattern("containerImage")
	if !ok {
		return fmt.Errorf("containerImage regex pattern not exist")
	}
	if !util.RegexStringChecker(c.Image, pattern) {
		return fmt.Errorf("container image invalid")
	}
	return nil
}

func (c *Container) checkContainerImageVersionValid() error {
	pattern, ok := appPattern.getPattern("imageVersion")
	if !ok {
		return fmt.Errorf("imageVersion regex pattern not exist")
	}
	if !util.RegexStringChecker(c.ImageVersion, pattern) {
		return fmt.Errorf("container image version invalid")
	}

	return nil
}

func (c *Container) checkContainerCommandValid() error {
	if len(c.Command) > commandMaxCount {
		return fmt.Errorf("container command count up to limt")
	}

	pattern, ok := appPattern.getPattern("containerCommand")
	if !ok {
		return fmt.Errorf("containerCommand regex pattern not exist")
	}

	for _, command := range c.Command {
		if !util.RegexStringChecker(command, pattern) {
			return fmt.Errorf("container command invalid")
		}
	}

	return nil
}

func (c *Container) checkContainerArgsValid() error {
	if len(c.Args) > argsMaxCount {
		return fmt.Errorf("container args count up to limt")
	}

	pattern, ok := appPattern.getPattern("containerArgs")
	if !ok {
		return fmt.Errorf("containerArgs regex pattern not exist")
	}

	for _, arg := range c.Args {
		if !util.RegexStringChecker(arg, pattern) {
			return fmt.Errorf("container arg invalid")
		}
	}

	return nil
}

func (c *Container) checkContainerEnvValid() error {
	if len(c.Env) > envMaxCount {
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
	for idx := range c.Env {
		if !util.RegexStringChecker(c.Env[idx].Name, namePattern) {
			return fmt.Errorf("container env var name invalid")
		}

		if !util.RegexStringChecker(c.Env[idx].Value, valuePattern) {
			return fmt.Errorf("container env var value invalid")
		}

		if _, ok := envNames[c.Env[idx].Name]; ok {
			return fmt.Errorf("container env value name is not unique")
		}
		envNames[c.Env[idx].Name] = struct{}{}
	}

	return nil
}

func (p *ContainerPort) checkPortName() error {
	pattern, ok := appPattern.getPattern("containerPortName")
	if !ok {
		return fmt.Errorf("containerPortName regex pattern not exist")
	}

	if !util.RegexStringChecker(p.Name, pattern) {
		return fmt.Errorf("container port name invalid")
	}
	return nil
}

func (p *ContainerPort) checkPortProtocol() error {
	if p.Proto != "TCP" && p.Proto != "UDP" {
		return fmt.Errorf("container port protocol invalid")
	}

	return nil
}

func (p *ContainerPort) checkPortContainerPort() error {
	if p.ContainerPort < minContainerPort || p.ContainerPort > maxContainerPort {
		return fmt.Errorf("container port invalid")
	}

	return nil
}

func (p *ContainerPort) checkPortHostPort() error {
	if p.HostPort < minHostPort || p.HostPort > maxHostPort {
		return fmt.Errorf("container host port invalid")
	}
	return nil
}

func (p *ContainerPort) checkPortHostIP() error {
	if p.HostIp == "" || p.HostIp == "0.0.0.0" || p.HostIp == "255.255.255.255" {
		return fmt.Errorf("container port host ip invalid")
	}

	ip := net.ParseIP(p.HostIp)
	if ip == nil || ip.To4() == nil {
		return fmt.Errorf("container port host ip is not ipv4")
	}

	return nil
}

func (c *Container) checkContainerPortsValid() error {
	if len(c.Ports) > portMapMaxCount {
		return fmt.Errorf("container ports num up to limit")
	}

	for _, port := range c.Ports {
		var checkItems = []func() error{
			port.checkPortName,
			port.checkPortProtocol,
			port.checkPortContainerPort,
			port.checkPortHostPort,
			port.checkPortHostIP,
		}
		for _, checkItem := range checkItems {
			if err := checkItem(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Container) checkUserIdValid() error {
	if c.UserId < minUserId || c.UserId > maxUserId {
		return fmt.Errorf("container user id valid")
	}

	return nil
}

func (c *Container) checkGroupIdValid() error {
	if c.UserId < minGroupId || c.UserId > maxGroupId {
		return fmt.Errorf("container group id valid")
	}

	return nil
}

func (a *AppParam) Check() error {
	var checkItems = []func() error{
		a.checkAppNameValid,
		a.checkAppDescriptionValid,
		a.checkContainerParaValid,
	}
	for _, checkItem := range checkItems {
		if err := checkItem(); err != nil {
			return err
		}
	}

	return nil
}
