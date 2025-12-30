// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package common to net utils
package common

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/checker/valid"
)

const (
	reservedIpv4Part0 = 10
)

// GetHostIpV4 get host ipv4
func GetHostIpV4() ([]net.IP, error) {
	allInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	ipv4 := make([]net.IP, 0)
	allAddr := make([]net.Addr, 0)
	for _, inter := range allInterfaces {
		if inter.Flags&net.FlagUp == 0 {
			continue
		}
		if strings.HasPrefix(inter.Name, "lo") {
			continue
		}
		addr, err := inter.Addrs()
		if err != nil {
			continue
		}
		allAddr = append(allAddr, addr...)
	}

	for _, addr := range allAddr {
		inet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if inet.IP.IsLoopback() {
			continue
		}

		ip := inet.IP.To4()
		if ip == nil {
			continue
		}
		ipv4 = append(ipv4, ip)
	}

	// allow both cluster ip and node ip for nginx manager
	nodeIpStr := os.Getenv("NODE_IP")
	if nodeIpStr == "" {
		return ipv4, nil
	}
	nodeIp := net.ParseIP(nodeIpStr)
	if nodeIp != nil {
		ipv4 = append(ipv4, nodeIp)
	}
	return ipv4, nil
}

// GetPodIP [method] for get pod ip from env
func GetPodIP() (string, error) {
	ip := os.Getenv("POD_IP")
	if ip == "" {
		return "", errors.New("pod ip obtained from env is nil")
	}

	if valid, err := valid.IsIpValid(ip); !valid {
		return "", fmt.Errorf("check pod ip failed: %v", err)
	}
	return ip, nil
}

// GetHostIP [method] for get host ip from env
func GetHostIP(envName string) (string, error) {
	ip := os.Getenv(envName)
	if ip == "" {
		return "", errors.New("host ip obtained from env is nil")
	}

	ipChecker := checker.GetIpV4Checker("", true)
	checkRes := ipChecker.Check(ip)
	if !checkRes.Result {
		return "", fmt.Errorf("host ip check failed: %s", checkRes.Reason)
	}
	return ip, nil
}
