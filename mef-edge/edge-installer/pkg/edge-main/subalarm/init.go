// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package subalarm this file for subAlarm manager module registers
package subalarm

import (
	"context"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/almutils"
	"edge-installer/pkg/common/constants"
)

// subAlarmMgr edge-main alarm module
type subAlarmMgr struct {
	ctx    context.Context
	enable bool
}

// NewSubAlarmModule new edge-main subAlarm module
func NewSubAlarmModule(ctx context.Context, enable bool) model.Module {
	module := &subAlarmMgr{
		enable: enable,
		ctx:    ctx,
	}
	return module
}

// Name module name
func (m *subAlarmMgr) Name() string {
	return constants.MainAlarmMgr
}

// Enable module enable
func (m *subAlarmMgr) Enable() bool {
	return m.enable
}

// Start module start running
func (m *subAlarmMgr) Start() {
	m.startMonitoring()
	for {
		select {
		case <-m.ctx.Done():
			return
		default:
		}
		msg, err := modulemgr.ReceiveMessage(m.Name())
		if err != nil {
			hwlog.RunLog.Errorf("get receive module message failed, error: %s", err.Error())
			continue
		}
		hwlog.RunLog.Infof("receive msg option: [%s] resource: [%s]", msg.GetOption(), msg.GetResource())
		m.dispatchMsg(msg)
	}
}

func (m *subAlarmMgr) dispatchMsg(msg *model.Message) {
	handlerMgr := GetHandlerMgr()
	if err := handlerMgr.Process(msg); err != nil {
		hwlog.RunLog.Errorf("handlers process failed, error: %v", err)
		return
	}
}

func (m *subAlarmMgr) startMonitoring() {
	var alarmMangerList []almutils.AlarmMonitor
	alarmMangerList = GetAlarmMonitorList()
	for _, alarm := range alarmMangerList {
		if alarm != nil {
			go alarm.Monitoring(m.ctx)
		}
	}
}
