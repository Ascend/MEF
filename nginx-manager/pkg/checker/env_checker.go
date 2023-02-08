// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package checker this file is for check parameter
package checker

import (
	"fmt"
	"strconv"

	"huawei.com/mindx/common/hwlog"

	"nginx-manager/pkg/nginxcom"
)

// EndpointTemplate nginx.conf需要替换的地址,端口项
var endpointTemplate = []nginxcom.Endpoint{
	{PortKey: nginxcom.EdgePortKey},
	{PortKey: nginxcom.SoftPortKey},
	{PortKey: nginxcom.CertPortKey},
	{PortKey: nginxcom.UserMgrSvcPortKey},
	{PortKey: nginxcom.NginxSslPortKey},
}

func checkEnv(envs interface{}) error {
	envMap, ok := envs.(map[string]string)
	if !ok {
		hwlog.RunLog.Error("env map type error")
		return fmt.Errorf("env map type error")
	}
	for _, v := range endpointTemplate {
		port, err := strconv.Atoi(envMap[v.PortKey])
		if err != nil {
			hwlog.RunLog.Errorf("port %d check fail", port)
			return fmt.Errorf("port %d check fail", port)
		}
		if port < nginxcom.PortMin || port > nginxcom.PortMax {
			hwlog.RunLog.Errorf("port %d check fail", port)
			return fmt.Errorf("port %d check fail", port)
		}
	}
	return nil
}
