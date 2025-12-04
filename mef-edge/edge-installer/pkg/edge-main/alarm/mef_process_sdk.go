// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package alarm this file for MEF alarm manager
package alarm

import (
	"context"
	"math/rand"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/almutils"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
)

const (
	dftReportAlarmInterval = 45 * time.Second
	reportMaxFluctuation   = 30
)

type alarmMEFManger struct {
	ctx context.Context
	alarmProcess
}

// NewAlarmMEFManager new alarm MEF manager
func NewAlarmMEFManager(ctx context.Context) ProxyManager {
	return &alarmMEFManger{
		ctx: ctx,
		alarmProcess: alarmProcess{
			processAlarmChan: make(chan EventAlarm, channelBufLen),
			processEventChan: make(chan EventAlarm),
			processing:       make(map[string]EventAlarm, defaultEventCapacity),
		},
	}
}

// StartMonitor prepare mef alarm and start monitor event
func (am *alarmMEFManger) StartMonitor() {
	var alarmIDs = map[string]struct{}{
		almutils.DockerAbnormal:  {},
		almutils.EdgeLogAbnormal: {},
		almutils.NPUAbnormal:     {},
		almutils.CertAbnormal:    {},
		almutils.EdgeDBAbnormal:  {},
	}
	am.loadAlarmFromDB(alarmIDs)
	go am.processAlarm(am.ctx)
	go am.processEvent(am.ctx)
	go am.scheduleReport(am.ctx)

}

func (am *alarmMEFManger) processEvent(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Warn("processEvent has stop")
			return
		case event, ok := <-am.processEventChan:
			if !ok {
				hwlog.RunLog.Error("processEventChan is closed")
				return
			}
			am.sendAlarm(event.Source, almutils.Alarms{
				Alarm: []almutils.Alarm{*event.Alarm},
			})
		}
	}
}

func (am *alarmMEFManger) scheduleReport(ctx context.Context) {
	interval := dftReportAlarmInterval + time.Duration(rand.Intn(reportMaxFluctuation))*time.Second
	tick := time.NewTicker(interval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Warn("report alarm to cloud regularly has stop")
			return
		case <-tick.C:
			am.sendAlarmToCloud()
			interval = dftReportAlarmInterval + time.Duration(rand.Intn(reportMaxFluctuation))*time.Second
			tick.Reset(interval)
		}
	}
}

// QueryAllAlarm process mef query alarm msg
func (am *alarmMEFManger) QueryAllAlarm(*model.Message) {
	return
}

func (am *alarmMEFManger) sendAlarmToCloud() {
	am.lock.RLock()
	defer am.lock.RUnlock()
	alarms := almutils.Alarms{
		Alarm: make([]almutils.Alarm, 0),
	}
	for _, event := range am.processing {
		am.setAlarmToDB(event)
		alarms.Alarm = append(alarms.Alarm, *event.Alarm)
	}
	am.sendAlarm(constants.AlarmManager, alarms)
}

func (am *alarmMEFManger) sendAlarm(source string, alarmSlice almutils.Alarms) {
	msg, err := util.NewInnerMsgWithFullParas(util.InnerMsgParams{
		Source:                source,
		Destination:           constants.ModEdgeHub,
		Operation:             constants.OptPost,
		Resource:              constants.ResMefAlarmReport,
		Content:               alarmSlice,
		TransferStructIntoStr: true,
	})
	if err != nil {
		hwlog.RunLog.Errorf("Create alarm msg failed, err: %s", err.Error())
		return
	}
	if err = modulemgr.SendMessage(msg); err != nil {
		hwlog.RunLog.Errorf("send async message failed, err is: %s", err.Error())
	}
}
