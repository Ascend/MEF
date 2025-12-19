// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
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
