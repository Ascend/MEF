// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager service
package appmanager

import (
	"fmt"
	"net"

	"edge-manager/pkg/util"
)

const (
	portMapMaxCount  = 16
	envMaxCount      = 256
	minContainerPort = 1
	maxContainerPort = 65535
	minHostPort      = 1024
	maxHostPort      = 65535
	minUserId        = 1
	maxUserId        = 65535
	minGroupId       = 1
	maxGroupId       = 65535
	commandMaxCount  = 16
	argsMaxCount     = 16
)

type containerParaChecker struct {
	container *util.Container
}

func (c *containerParaChecker) checkContainerNameValid() error {
	if !util.RegexStringChecker(c.container.Name, "^[a-z0-9]([a-z0-9-]{0,30}[a-z0-9]){0,1}$") {
		return fmt.Errorf("container name invalid")
	}
	return nil
}

func (c *containerParaChecker) checkContainerImageValid() error {
	if !util.RegexStringChecker(c.container.Image, "^[a-z]([a-z0-9_./-]{0,30}[a-z0-9]){0,1}$") {
		return fmt.Errorf("container image invalid")
	}
	return nil
}

func (c *containerParaChecker) checkContainerImageVersionValid() error {
	if !util.RegexStringChecker(c.container.ImageVersion, "^[a-zA-Z0-9_.-]{1,32}$") {
		return fmt.Errorf("container image version invalid")
	}

	return nil
}

func (c *containerParaChecker) checkContainerCommandValid() error {
	if len(c.container.Command) > commandMaxCount {
		return fmt.Errorf("container command count up to limt")
	}

	for _, command := range c.container.Command {
		if !util.RegexStringChecker(command, "^[a-zA-Z0-9 _./-]{0,31}[a-zA-Z0-9]$") {
			return fmt.Errorf("container command invalid")
		}
	}

	return nil
}

func (c *containerParaChecker) checkContainerArgsValid() error {
	if len(c.container.Args) > argsMaxCount {
		return fmt.Errorf("container args count up to limt")
	}

	for _, arg := range c.container.Args {
		if !util.RegexStringChecker(arg, "^[a-zA-Z0-9 _./-]{0,31}[a-zA-Z0-9]$") {
			return fmt.Errorf("container arg invalid")
		}
	}

	return nil
}

func (c *containerParaChecker) checkContainerEnvValid() error {
	if len(c.container.Env) > envMaxCount {
		return fmt.Errorf("container image env var num up to limit")
	}
	var envNames = map[string]struct{}{}
	for idx := range c.container.Env {
		if !util.RegexStringChecker(c.container.Env[idx].Name, "^[a-zA-Z][a-zA-z0-9._-]{0,30}[a-zA-Z0-9]$") {
			return fmt.Errorf("container env var name invalid")
		}

		if !util.RegexStringChecker(c.container.Env[idx].Value, "^[a-zA-Z0-9 _./-]{0,512}$") {
			return fmt.Errorf("container env var value invalid")
		}

		if _, ok := envNames[c.container.Env[idx].Name]; ok {
			return fmt.Errorf("container env value name is not unique")
		}
		envNames[c.container.Env[idx].Name] = struct{}{}
	}

	return nil
}

type portParaChecker struct {
	port *util.ContainerPort
}

func (c *portParaChecker) checkPortName() error {
	if !util.RegexStringChecker(c.port.Name, "^[a-z0-9]([a-z0-9-]{0,30}[a-z0-9]){0,1}$") {
		return fmt.Errorf("container port name invalid")
	}
	return nil
}

func (c *portParaChecker) checkPortProtocol() error {
	if c.port.Proto != "TCP" && c.port.Proto != "UDP" {
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

type appCreatParaChecker struct {
	req *util.CreateAppReq
}

func (c *appCreatParaChecker) checkAppNameValid() error {
	if !util.RegexStringChecker(c.req.AppName, "^[a-z]([a-z0-9-]{0,30}[a-z0-9]){0,1}$") {
		return fmt.Errorf("app name invalid")
	}
	return nil
}

func (c *appCreatParaChecker) checkAppVersionValid() error {
	if !util.RegexStringChecker(c.req.Version, "^[a-z0-9][a-z0-9.]{0,6}[a-z0-9]$") {
		return fmt.Errorf("app version invalid")
	}

	return nil
}

func (c *appCreatParaChecker) checkAppDescriptionValid() error {
	if !util.RegexStringChecker(c.req.Description, `^[\S]{0,512}$`) {
		return fmt.Errorf("app description invalid")
	}
	return nil
}

func (c *appCreatParaChecker) checkAppContainersValid() error {
	for idx := range c.req.Containers {
		var checker = containerParaChecker{container: &c.req.Containers[idx]}
		var checkItems = []func() error{
			checker.checkContainerNameValid,
			checker.checkContainerImageValid,
			checker.checkContainerImageVersionValid,
			checker.checkContainerCommandValid,
			checker.checkContainerArgsValid,
			checker.checkContainerEnvValid,
			checker.checkContainerPortsValid,
			checker.checkUserIdValid,
			checker.checkGroupIdValid,
		}
		for _, checkItem := range checkItems {
			if err := checkItem(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *appCreatParaChecker) Check() error {
	var checkItems = []func() error{
		c.checkAppNameValid,
		c.checkAppVersionValid,
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

type appDeployParaChecker struct {
	req *util.DeployAppReq
}

func (c *appDeployParaChecker) checkNodeGroupNameValid() error {
	if !util.RegexStringChecker(c.req.NodeGroupName, "^[a-z]([a-z0-9-]{0,30}[a-z0-9]){0,1}$") {
		return fmt.Errorf("container name invalid")
	}
	return nil
}

func (c *appDeployParaChecker) Check() error {
	var checkItems = []func() error{
		c.checkNodeGroupNameValid,
	}
	for _, checkItem := range checkItems {
		if err := checkItem(); err != nil {
			return err
		}
	}
	return nil
}
