// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build !MEFEdge_SDK

// Package subalarm get monitor list
package subalarm

import "edge-installer/pkg/common/almutils"

// GetAlarmMonitorList return edge-main alarm monitor list
func GetAlarmMonitorList() []almutils.AlarmMonitor {
	return []almutils.AlarmMonitor{}
}
