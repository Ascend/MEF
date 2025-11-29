// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package subalarm get monitor list
package subalarm

import (
	"time"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/almutils"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common/configpara"
	"edge-installer/pkg/edge-main/subalarm/monitors"
)

// GetAlarmMonitorList return edge-main alarm monitor list
func GetAlarmMonitorList() []almutils.AlarmMonitor {
	for i := 0; i < constants.TryConnectNet; i++ {
		netTypeStr, err := configpara.GetNetType()
		if err != nil {
			time.Sleep(constants.StartWsWaitTime)
			continue
		}

		switch netTypeStr {
		case constants.MEF:
			return monitors.GetMEFMonitorList()
		default:
			time.Sleep(constants.StartWsWaitTime)
			continue
		}
	}

	hwlog.RunLog.Error("get net type failed, reached the maximum number of the connection attempts")
	return []almutils.AlarmMonitor{}
}
