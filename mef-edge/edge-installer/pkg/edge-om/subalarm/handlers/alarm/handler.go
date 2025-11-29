// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package alarm this file for alarm-mgr handler
package alarm

import (
	"errors"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/almutils"
	"edge-installer/pkg/edge-om/subalarm/handlers/monitors"
)

var alarmMangerList []almutils.AlarmMonitor

// Handler alarm connect handler
type Handler struct {
}

func init() {
	alarmMangerList = monitors.GetAlarmManagerList()
}

// Handle configHandler handle entry
func (ch *Handler) Handle(msg *model.Message) error {
	hwlog.RunLog.Info("start to report all alarm msg")
	var report bool
	if err := msg.ParseContent(&report); err != nil {
		hwlog.RunLog.Errorf("get report content failed: %v", err)
		return errors.New("get report content failed")
	}

	hwlog.RunLog.Infof("report alarm state is :%t", report)
	if !report {
		hwlog.RunLog.Error("not report alarm")
		return errors.New("not report alarm")
	}
	for _, alarm := range alarmMangerList {
		if alarm != nil {
			alarm.CollectOnce()
		}
	}
	return nil
}
