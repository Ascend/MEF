// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package alarm this file for FD alarm manager
package alarm

import (
	"context"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/almutils"
	"edge-installer/pkg/common/constants"
)

type alarmFDManger struct {
	ctx context.Context
	alarmProcess
}

// NewAlarmFDManager new alarm FD manager
func NewAlarmFDManager(ctx context.Context) ProxyManager {
	return &alarmFDManger{
		ctx: ctx,
		alarmProcess: alarmProcess{
			processAlarmChan: make(chan EventAlarm, channelBufLen),
			processEventChan: make(chan EventAlarm),
			processing:       make(map[string]EventAlarm, defaultEventCapacity),
		},
	}
}

// StartMonitor prepare fd alarm and start monitor event
func (am *alarmFDManger) StartMonitor() {
	var alarmIDs = map[string]struct{}{
		almutils.DockerAbnormal:  {},
		almutils.EdgeLogAbnormal: {},
		almutils.NPUAbnormal:     {},
		almutils.EdgeDBAbnormal:  {},
	}
	am.loadAlarmFromDB(alarmIDs)
	go am.processAlarm(am.ctx)
	go am.processEvent(am.ctx)
}

func (am *alarmFDManger) processEvent(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Warn("processEvent has stop")
			return
		case event, ok := <-am.processEventChan:
			if !ok {
				hwlog.RunLog.Error("The processEventChan is closed")
				return
			}
			if err := almutils.SendAlarm(event.Source, constants.ModDeviceOm, event.Alarm); err != nil {
				hwlog.RunLog.Errorf("send alarm to cloud failed, error:%s", err.Error())
			}
		}
	}
}

// QueryAllAlarm process fd query alarm msg
func (am *alarmFDManger) QueryAllAlarm(*model.Message) {
	am.lock.Lock()
	defer am.lock.Unlock()
	alarms := almutils.Alarms{
		Alarm: make([]almutils.Alarm, 0),
	}
	for _, event := range am.processing {
		am.setAlarmToDB(event)
		alarms.Alarm = append(alarms.Alarm, *event.Alarm)
	}
	am.sendAlarmToCloud(alarms)
}

func (am *alarmFDManger) sendAlarmToCloud(alarms almutils.Alarms) {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("Create alarm msg failed, err: %s", err.Error())
		return
	}
	msg.Header.ID = msg.Header.Id
	msg.SetRouter(constants.AlarmManager, constants.ModDeviceOm, constants.OptResponse, constants.QueryAllAlarm)
	msg.SetKubeEdgeRouter(
		constants.AlarmManager, constants.ModDeviceOm, constants.OptResponse, constants.QueryAllAlarm)
	if err = msg.FillContent(alarms, true); err != nil {
		hwlog.RunLog.Errorf("fill alarms into content failed: %v", err)
		return
	}
	if err = modulemgr.SendAsyncMessage(msg); err != nil {
		hwlog.RunLog.Errorf("send async message failed, err is: %s", err.Error())
	}
}
