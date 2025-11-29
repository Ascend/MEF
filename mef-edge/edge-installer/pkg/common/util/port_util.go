// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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
