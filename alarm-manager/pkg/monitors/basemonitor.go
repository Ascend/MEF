// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package monitors defined base monitor
package monitors

import (
	"context"
	"time"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common/alarms"
	"huawei.com/mindxedge/base/common/requests"
)

// AlarmMonitor alarm report interface
type AlarmMonitor interface {
	Monitoring(ctx context.Context)
	CollectOnce()
}

type cronTask struct {
	name           string
	interval       time.Duration
	alarmIdFuncMap map[string]func() error
	resetFunc      func()
}

// Monitoring monitor one task and call collectOnce
func (ct *cronTask) Monitoring(ctx context.Context) {
	ct.CollectOnce()

	tick := time.NewTicker(ct.interval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Warnf("monitor %s stop", ct.name)
			return
		case <-tick.C:
			if ct.resetFunc != nil {
				ct.resetFunc()
			}
			ct.CollectOnce()
		}
	}
}

// CollectOnce call check func and send alarm req
func (ct *cronTask) CollectOnce() {
	if ct.alarmIdFuncMap == nil {
		return
	}
	if len(ct.alarmIdFuncMap) == 0 {
		return
	}

	var alarmReqs []*requests.AlarmReq
	var notifyType string
	for alarmId, checkFunc := range ct.alarmIdFuncMap {
		notifyType = alarms.ClearFlag
		if err := checkFunc(); err != nil {
			hwlog.RunLog.Warnf("%s task check failed, error: %v", ct.name, err)
			notifyType = alarms.AlarmFlag
		}

		alarmReq, err := alarms.CreateAlarm(alarmId, ct.name, notifyType)
		if err != nil {
			hwlog.RunLog.Errorf("create alarm %s failed, error: %v", alarmId, err)
			continue
		}
		alarmReqs = append(alarmReqs, alarmReq)
	}

	if err := SendAlarms(alarmReqs...); err != nil {
		hwlog.RunLog.Errorf("send alarms failed, error: %v", err)
		return
	}

	hwlog.RunLog.Infof("send %s alarm success", ct.name)
	return
}
