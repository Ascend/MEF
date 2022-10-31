// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package common for parameter validate
package common

import (
	"errors"
	"net"

	"huawei.com/mindx/common/hwlog"
)

// BaseParamValid doing param check
func BaseParamValid(port int, ip string) error {
	if port < minPort || port > maxPort {
		return errors.New("the port is invalid")
	}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return errors.New("the listen ip is invalid")
	}
	ip = parsedIP.String()
	hwlog.RunLog.Infof("listen on: %s", ip)

	return nil
}
