// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package util

import (
	"path/filepath"
	"strconv"
	"strings"

	"huawei.com/mindx/common/fileutils"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// GetUsedPorts get ports of TCP protocol
func GetUsedPorts(protocol v1.Protocol) (sets.Int64, error) {
	const (
		sysNetFilePrefix   = "/proc/net/"
		netFileValidMinLen = 2
		localAddressLen    = 2

		hexadecimalBase = 16
		intBit          = 64
	)

	usedPorts := sets.NewInt64()
	protocolStatFile := sysNetFilePrefix + strings.ToLower(string(protocol))
	realPath, err := filepath.EvalSymlinks(protocolStatFile)
	if err != nil {
		return nil, err
	}
	data, err := fileutils.LoadFile(realPath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")

	for i := 1; i < len(lines); i++ {
		fields := strings.Fields(lines[i])
		if len(fields) < netFileValidMinLen {
			continue
		}
		ipPort := strings.Split(fields[1], ":")
		if len(ipPort) != localAddressLen {
			continue
		}
		port, err := strconv.ParseInt(ipPort[1], hexadecimalBase, intBit)
		if err != nil {
			continue
		}
		usedPorts.Insert(port)
	}
	return usedPorts, nil
}
