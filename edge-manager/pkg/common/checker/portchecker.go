// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package checker to check port
package checker

// IsPortInRange check port is in range or not
func IsPortInRange(minPort, maxPort, port int) bool {
	if port < minPort || port > maxPort {
		return false
	}
	return true
}
