// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package almutils
package almutils

import (
	"errors"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
)

// CreateAlarm creates an alarm
func CreateAlarm(alarmId, resource, notifyType string) (*Alarm, error) {
	template, ok := idToAlarms[alarmId]
	if !ok {
		hwlog.RunLog.Errorf("unknown alarm type, alarm id [%s]", alarmId)
		return nil, errors.New("unknown alarm type")
	}
	alm := template
	alm.Resource = resource
	alm.Timestamp = time.Now().Format(time.RFC3339)
	alm.NotificationType = notifyType
	return &alm, nil
}

// SendAlarm sends alarms
func SendAlarm(source, destination string, alarm ...*Alarm) error {
	if len(alarm) == 0 {
		return errors.New("alarm is required")
	}
	var alarmSlice []Alarm
	for _, alm := range alarm {
		if alm == nil {
			return errors.New("alarm can't be nil pointer")
		}
		alarmSlice = append(alarmSlice, *alm)
	}
	alarms := Alarms{
		Alarm: alarmSlice,
	}
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("Create alarm msg failed, err: %v", err)
		return err
	}
	msg.Header.ID = msg.Header.Id
	msg.SetKubeEdgeRouter(constants.EdgedModule, constants.GroupHub, constants.OptUpdate, constants.ResAlarm)
	msg.SetRouter(source, destination, constants.OptUpdate, constants.ResAlarm)
	if err = msg.FillContent(alarms, true); err != nil {
		hwlog.RunLog.Errorf("fill alarm content failed: %v", err)
		return errors.New("fill alarm content failed")
	}
	return modulemgr.SendAsyncMessage(msg)
}

// CreateAndSendAlarm create and send alarm
func CreateAndSendAlarm(alarmId, resource, notifyType, source, destination string) error {
	createdAlarm, err := CreateAlarm(alarmId, resource, notifyType)
	if err != nil {
		return err
	}
	return SendAlarm(source, destination, createdAlarm)
}
