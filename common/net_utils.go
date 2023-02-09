// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package common to net utils
package common

import (
	"net"
	"os"
	"strings"
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
		if ip[0] != reservedIpv4Part0 && inet.IP.IsPrivate() {
			continue
		}
		ipv4 = append(ipv4, ip)
	}

	// for k8s pod environment
	if len(ipv4) == 0 {
		nodeIpStr := os.Getenv("NODE_IP")
		if nodeIpStr == "" {
			return ipv4, nil
		}
		nodeIp := net.ParseIP(nodeIpStr)
		if nodeIp != nil {
			ipv4 = append(ipv4, nodeIp)
		}
	}
	return ipv4, nil
}
