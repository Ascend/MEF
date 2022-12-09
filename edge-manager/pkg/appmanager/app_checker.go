// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager service
package appmanager

import (
	"edge-manager/pkg/util"
	"fmt"
)

type appParaPattern struct {
	patterns map[string]string
}

var appPattern = appParaPattern{patterns: map[string]string{
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

func (a *appParaPattern) getPattern(key string) (string, bool) {
	pattern, ok := a.patterns[key]
	return pattern, ok
}

func (c *CreateAppReq) Check() error {
	return c.AppParam.Check()
}

func (c *DeployAppReq) checkNodeGroupNameValid() error {
	pattern, ok := appPattern.getPattern("nodeGroupName")
	if !ok {
		return fmt.Errorf("nodeGroupName regex pattern not exist")
	}

	for _, nodeGroupInfo := range c.NodeGroupInfo {
		if !util.RegexStringChecker(nodeGroupInfo.NodeGroupName, pattern) {
			return fmt.Errorf("container name invalid")
		}
	}

	return nil
}

func (c *DeployAppReq) Check() error {
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
