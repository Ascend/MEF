// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package checker this file is for check parameter
package checker

import (
	"fmt"
	"strconv"

	"huawei.com/mindxedge/base/common/checker"
	"nginx-manager/pkg/nginxcom"

	"huawei.com/mindx/common/hwlog"
)

// EndpointTemplate the ports needed to replace int nginx.conf
var envPorts = []nginxcom.EnvEntry{
	{EnvKey: nginxcom.EdgePortKey},
	{EnvKey: nginxcom.SoftPortKey},
	{EnvKey: nginxcom.CertPortKey},
	{EnvKey: nginxcom.UserMgrSvcPortKey},
	{EnvKey: nginxcom.NginxSslPortKey},
}

var envIps = []nginxcom.EnvEntry{
	{EnvKey: nginxcom.PodIpKey},
}

func checkEnv(envs interface{}) error {
	envMap, ok := envs.(map[string]string)
	if !ok {
		hwlog.RunLog.Error("env map type error")
		return fmt.Errorf("env map type error")
	}
	err := checkPorts(envMap)
	if err != nil {
		return err
	}
	return checkIps(envMap)
}

func checkPorts(envMap map[string]string) error {
	for _, v := range envPorts {
		port, err := strconv.Atoi(envMap[v.EnvKey])
		if err != nil {
			hwlog.RunLog.Errorf("port %d check fail", port)
			return fmt.Errorf("port %d check fail", port)
		}
		if !checker.IsPortInRange(nginxcom.PortMin, nginxcom.PortMax, port) {
			hwlog.RunLog.Errorf("port %d check fail", port)
			return fmt.Errorf("port %d check fail", port)
		}
	}
	return nil
}

func checkIps(envMap map[string]string) error {
	for _, v := range envIps {
		ip, ok := envMap[v.EnvKey]
		if !ok {
			hwlog.RunLog.Errorf("env %s not exist", v.EnvKey)
			return fmt.Errorf("env %s not exist", v.EnvKey)
		}
		valid, _ := checker.IsIpValid(ip)
		if !valid {
			hwlog.RunLog.Errorf("ip %s check fail", ip)
			return fmt.Errorf("ip %s check fail", ip)
		}
	}
	return nil
}
