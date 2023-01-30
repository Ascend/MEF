// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package checker this file is for check parameter
package checker

import (
	"fmt"
	"strconv"

	"nginx-manager/pkg/nginxcom"
)

// EndpointTemplate nginx.conf需要替换的地址,端口项
var endpointTemplate = []nginxcom.Endpoint{
	{UrlKey: nginxcom.EdgeUrlKey, PortKey: nginxcom.EdgePortKey, Regexp: nginxcom.UrlPattern},
	{UrlKey: nginxcom.SoftUrlKey, PortKey: nginxcom.SoftPortKey, Regexp: nginxcom.UrlPattern},
	{UrlKey: nginxcom.CertUrlKey, PortKey: nginxcom.CertPortKey, Regexp: nginxcom.UrlPattern},
}

func checkEnv(envs interface{}) error {
	envMap, ok := envs.(map[string]string)
	if !ok {
		return fmt.Errorf("env map type error")
	}
	for _, v := range endpointTemplate {
		if _, ok := envMap[v.UrlKey]; !ok {
			return fmt.Errorf("no env %s found", v.UrlKey)
		}
		if _, ok := envMap[v.PortKey]; !ok {
			return fmt.Errorf("no env %s found", v.UrlKey)
		}
		if !RegexStringChecker(envMap[v.UrlKey], v.Regexp) {
			return fmt.Errorf("env %s pattern check fail", v.UrlKey)
		}
		port, err := strconv.Atoi(envMap[v.PortKey])
		if err != nil {
			return fmt.Errorf("port %d check fail", port)
		}
		if port < nginxcom.PortMin || port > nginxcom.PortMax {
			return fmt.Errorf("port %d check fail", port)
		}
	}
	return nil
}
