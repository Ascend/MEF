// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package common get/set fd ip for edge-main proc
package common

import (
	"fmt"
	"strings"
	"sync"

	"edge-installer/pkg/common/constants"
)

var (
	fdIp       string
	ipRwLocker sync.Mutex
)

const maxAddrLen = 256
const minimalAddrDataLen = 2

// SetFdIp set fd ip, parameter is ip:port format
func SetFdIp(modName string, addr string) error {
	ipRwLocker.Lock()
	defer ipRwLocker.Unlock()
	if len(addr) > maxAddrLen {
		return fmt.Errorf("invalid addr: %v", addr)
	}
	ipPort := strings.Split(addr, ":")
	if len(ipPort) < minimalAddrDataLen {
		return fmt.Errorf("invalid addr: %v", addr)
	}
	switch modName {
	case constants.ModDeviceOm:
		fdIp = ipPort[0]
	default:
		return fmt.Errorf("set fd ip failed. invalid module name: %v", modName)
	}
	return nil
}

// GetFdIp get fd ip
func GetFdIp() (string, error) {
	ipRwLocker.Lock()
	defer ipRwLocker.Unlock()
	if fdIp == "" {
		return "", fmt.Errorf("fd ip not found")
	}
	return fdIp, nil
}
