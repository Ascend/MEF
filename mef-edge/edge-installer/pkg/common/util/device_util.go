// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package util this file for getting device information
package util

import (
	"errors"
	"fmt"
	"strings"

	"huawei.com/mindx/common/envutils"

	"edge-installer/pkg/common/constants"
)

const (
	eLabel   = "elabel"
	get      = "get"
	a500SnId = "0x34"
	snIndex  = 1
)

const (
	info            = "info"
	format          = "--format"
	cgroupDriverFmt = "{{.CgroupDriver}}"
)

func getA500Sn() (string, error) {
	output, err := envutils.RunCommand(eLabel, envutils.DefCmdTimeoutSec, get, a500SnId)
	if err != nil {
		return "", fmt.Errorf("get a500 serial number failed,error:%v", err)
	}
	arr := strings.Split(output, " ")
	if len(arr) > snIndex {
		sn := strings.TrimRight(arr[snIndex], "\r\n")
		return sn, nil
	}
	return "", errors.New("get a500 serial number failed,error:output is not expected format")
}

// GetCgroupDriver get cgroup driver
func GetCgroupDriver() (string, error) {
	return envutils.RunCommand(constants.DockerCmd, envutils.DefCmdTimeoutSec, info, format, cgroupDriverFmt)
}
