// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package monitors for base task monitor
package monitors

import (
	"context"
	"time"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/almutils"
	"edge-installer/pkg/common/constants"
)

type cronTask struct {
	alarmId         string
	name            string
	interval        time.Duration
	checkStatusFunc func() error
}

// GetAlarmManagerList return edge-om alarm manager list
func GetAlarmManagerList() []almutils.AlarmMonitor {
	return []almutils.AlarmMonitor{
		logTask,
		dockerTask,
		npuTask,
		dbTask,
	}
}

// Monitoring monitor one task and call collectOnce
func (ct *cronTask) Monitoring(ctx context.Context) {
	tick := time.NewTicker(ct.interval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Warnf("monitor %s stop", ct.name)
			return
		case <-tick.C:
			ct.CollectOnce()
			tick.Reset(ct.interval)
		}
	}
}

// CollectOnce call check func and send alarm
func (ct *cronTask) CollectOnce() {
	if ct.checkStatusFunc == nil {
		return
	}
	notifyType := almutils.NotifyTypeClear
	if err := ct.checkStatusFunc(); err != nil {
		hwlog.RunLog.Warnf("%s task check failed, err: %v", ct.name, err)
		notifyType = almutils.NotifyTypeAlarm
	}

	if err := almutils.CreateAndSendAlarm(
		ct.alarmId, ct.name, notifyType, ct.name, constants.InnerClient); err != nil {
		hwlog.RunLog.Errorf("send alarm failed, %v", err)
		return
	}
	hwlog.RunLog.Infof("send %s alarm (notifyType=%s) success", ct.name, notifyType)
}
