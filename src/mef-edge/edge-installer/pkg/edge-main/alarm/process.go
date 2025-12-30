// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package alarm this file for base alarm process
package alarm

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/almutils"
	"edge-installer/pkg/edge-main/msgconv/statusmanager"
)

const (
	defaultEventCapacity = 16
	channelBufLen        = 1024
)

// ProxyManager proxy manager interface
type ProxyManager interface {
	StartMonitor()
	UpdateAlarm(*model.Message)
	QueryAllAlarm(*model.Message)
}

// EventAlarm event alarm module
type EventAlarm struct {
	Source string
	Alarm  *almutils.Alarm
}

type alarmProcess struct {
	lock             sync.RWMutex
	processEventChan chan EventAlarm
	processAlarmChan chan EventAlarm
	processing       map[string]EventAlarm
}

func (am *alarmProcess) loadAlarmFromDB(alarmIDs map[string]struct{}) {
	result, err := statusmanager.GetAlarmStatusMgr().GetAll()
	if err != nil {
		hwlog.RunLog.Errorf("Get alarm from db error:%s", err.Error())
		return
	}
	am.lock.Lock()
	defer am.lock.Unlock()
	for _, eventStr := range result {
		var event EventAlarm
		if err = json.Unmarshal([]byte(eventStr), &event); err != nil {
			hwlog.RunLog.Errorf("Unmarshal alarm events failed, %s", err.Error())
			continue
		}
		if _, ok := alarmIDs[event.Alarm.AlarmId]; !ok ||
			event.Alarm.NotificationType == almutils.NotifyTypeClear {
			am.deleteAlarmToDB(event)
			continue
		}
		am.processing[event.Alarm.AlarmId] = event
	}
}

func (am *alarmProcess) processAlarm(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Warn("processAlarm has stop")
			return
		case alarm, ok := <-am.processAlarmChan:
			if !ok {
				hwlog.RunLog.Error("processAlarmChan is closed")
				return
			}
			am.process(alarm)
		}
	}
}

func (am *alarmProcess) process(alarm EventAlarm) {
	am.lock.Lock()
	defer am.lock.Unlock()
	am.processing[alarm.Alarm.AlarmId] = alarm
	am.setAlarmToDB(alarm)
}

// UpdateAlarm process update alarm msg
func (am *alarmProcess) UpdateAlarm(msg *model.Message) {
	var alarms almutils.Alarms
	if err := msg.ParseContent(&alarms); err != nil {
		hwlog.RunLog.Errorf("parse alarm msg failed, error: %s", err.Error())
		return
	}
	for _, alm := range alarms.Alarm {
		am.handleAlarm(msg.Router.Source, &alm)
	}
}

func (am *alarmProcess) filterAlarm(alm *almutils.Alarm) bool {
	am.lock.RLock()
	defer am.lock.RUnlock()
	oldEvent, ok := am.processing[alm.AlarmId]
	if !ok && alm.NotificationType == almutils.NotifyTypeClear {
		return true
	}
	if ok && oldEvent.Alarm.NotificationType == alm.NotificationType {
		return true
	}
	return false
}

func (am *alarmProcess) handleAlarm(almSource string, alm *almutils.Alarm) {
	event := EventAlarm{
		Alarm:  alm,
		Source: almSource,
	}
	if alm.Type == almutils.TypeEvent {
		am.processEventChan <- event
		return
	}
	if am.filterAlarm(alm) {
		hwlog.RunLog.Warnf("duplicate or ignore alarm [%s, %s], ignore.",
			alm.NotificationType, alm.AlarmId)
		return
	}
	am.processAlarmChan <- event
}

func (am *alarmProcess) setAlarmToDB(event EventAlarm) {
	if !strings.HasPrefix(event.Alarm.AlarmId, almutils.MefAlarmIdPrefix) {
		return
	}
	// Alarms will be stored in database
	if err := statusmanager.GetAlarmStatusMgr().Set(event.Alarm.AlarmId, event); err != nil {
		hwlog.RunLog.Errorf("failed to save alarm to db, %s", err.Error())
	}
	return
}

func (am *alarmProcess) deleteAlarmToDB(event EventAlarm) {
	if !strings.HasPrefix(event.Alarm.AlarmId, almutils.MefAlarmIdPrefix) {
		return
	}
	// Alarms will be delete in database
	if err := statusmanager.GetAlarmStatusMgr().Delete(event.Alarm.AlarmId); err != nil {
		hwlog.RunLog.Errorf("failed to delete alarm to db, %s", err.Error())
	}
	return
}
