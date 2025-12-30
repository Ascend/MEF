// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package subalarm this file for subAlarm manager module registers
package subalarm

import (
	"context"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/almutils"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-om/subalarm/handlers"
	"edge-installer/pkg/edge-om/subalarm/handlers/monitors"
)

var alarmMangerList []almutils.AlarmMonitor

// subAlarmMgr edge-om alarm module
type subAlarmMgr struct {
	ctx    context.Context
	cancel context.CancelFunc
	enable bool
}

func init() {
	alarmMangerList = monitors.GetAlarmManagerList()
}

// NewSubAlarmModule new edge-om subAlarm module
func NewSubAlarmModule(enable bool) model.Module {
	module := &subAlarmMgr{
		enable: enable,
	}
	module.ctx, module.cancel = context.WithCancel(context.Background())
	return module
}

// Name module name
func (m *subAlarmMgr) Name() string {
	return constants.OmAlarmMgr
}

// Enable module enable
func (m *subAlarmMgr) Enable() bool {
	return m.enable
}

// Stop module stop
func (m *subAlarmMgr) Stop() bool {
	m.cancel()
	return true
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
			hwlog.RunLog.Errorf("get receive module message failed,error:%s", err.Error())
			continue
		}
		hwlog.RunLog.Infof("receive msg option:[%s] resource:[%s]", msg.GetOption(), msg.GetResource())
		go m.dispatchMsg(msg)
	}
}

func (m *subAlarmMgr) dispatchMsg(msg *model.Message) {
	handlerMgr := handlers.GetHandlerMgr()
	if err := handlerMgr.Process(msg); err != nil {
		hwlog.RunLog.Errorf("handlers process failed, err:%s", err.Error())
		return
	}
}

func (m *subAlarmMgr) startMonitoring() {
	for _, alarm := range alarmMangerList {
		if alarm != nil {
			go alarm.Monitoring(m.ctx)
		}
	}
}
