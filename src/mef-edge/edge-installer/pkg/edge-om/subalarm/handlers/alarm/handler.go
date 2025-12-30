// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
