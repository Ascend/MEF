// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package checker to check ip
package valid

import (
	"errors"
	"net"
)

// IsIpValid check ip is valid
func IsIpValid(ip string) (bool, error) {
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		return false, errors.New("ip parse failed")
	}
	if parsedIp.To4() == nil && parsedIp.To16() == nil {
		return false, errors.New("IP must be a valid IP address")
	}
	if parsedIp.Equal(net.IPv4bcast) {
		return false, errors.New("IP can't be a broadcast address")
	}
	if parsedIp.IsMulticast() {
		return false, errors.New("IP can't be a multicast address")
	}
	if parsedIp.IsLinkLocalUnicast() {
		return false, errors.New("IP can't be a link-local unicast address")
	}
	if parsedIp.IsUnspecified() {
		return false, errors.New("IP can't be an all zeros address")
	}
	return true, nil
}

// IsPortInRange check port is in range or not
func IsPortInRange(minPort, maxPort, port int) bool {
	return port >= minPort && port <= maxPort
}
