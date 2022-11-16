// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package checker to check ip
package checker

import (
	"fmt"
	"huawei.com/mindx/common/hwlog"
	"net"
)

// IsIpValid check ip is valid
func IsIpValid(ip string) (bool, error) {
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		return false, fmt.Errorf("ip parse failed")
	}
	if parsedIp.To4() == nil && parsedIp.To16() == nil {
		return false, fmt.Errorf("IP must be a valid IP address")
	}
	if parsedIp.Equal(net.IPv4bcast) {
		return false, fmt.Errorf("IP can't be a broadcast address")
	}
	if parsedIp.IsMulticast() {
		return false, fmt.Errorf("IP can't be a multicast address")
	}
	if parsedIp.IsLinkLocalUnicast() {
		return false, fmt.Errorf("IP can't be a link-local unicast address")
	}
	if parsedIp.IsUnspecified() {
		return false, fmt.Errorf("IP can't be an all zeros address")
	}
	return true, nil
}

// IsIpInHost check whether the IP address is on the host
func IsIpInHost(ip string) (bool, error) {
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		return false, fmt.Errorf("ip parse failed")
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		hwlog.RunLog.Errorf("get host ip list fail")
		return false, fmt.Errorf("get host ip list fail")
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		default:
			hwlog.RunLog.Error("unexpected type %T", v)
		}
		if ip != nil && ip.Equal(parsedIp) {
			return true, nil
		}
	}
	return false, fmt.Errorf("ip %s not found in the host's network interfaces", ip)
}
