// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package alarm this file for register alarm manager
package alarm

import (
	"context"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
)

// alarmManager alarm module
type alarmManager struct {
	ctx         context.Context
	enable      bool
	proxyManger ProxyManager
}

// NewAlarmManager new alarm manager module
func NewAlarmManager(ctx context.Context, enable bool) model.Module {
	return &alarmManager{
		enable: enable,
		ctx:    ctx,
	}
}

// Name module name
func (am *alarmManager) Name() string {
	return constants.AlarmManager
}

// Enable module enable
func (am *alarmManager) Enable() bool {
	if err := am.registerManager(); err != nil {
		hwlog.RunLog.Errorf("register failed, err:%s", err.Error())
		return false
	}
	return am.enable
}

// Start module start running
func (am *alarmManager) Start() {
	if am.proxyManger == nil {
		hwlog.RunLog.Error("register proxy manager failed")
		return
	}
	am.proxyManger.StartMonitor()
	am.runAlarmManager()
}

// runAlarmManager is received alarm message
func (am *alarmManager) runAlarmManager() {
	hwlog.RunLog.Info("alarm manager start success")
	for {
		select {
		case <-am.ctx.Done():
			hwlog.RunLog.Info("-------------------alarm manager exit-------------------")
			return
		default:
		}
		msg, err := modulemgr.ReceiveMessage(am.Name())
		if err != nil {
			hwlog.RunLog.Errorf("alarm manager receive module message failed, error: %v", err)
			continue
		}
		am.processAlarmMessage(msg)
	}
}

func (am *alarmManager) processAlarmMessage(msg *model.Message) {
	registerInfoList := []struct {
		operation   string
		resource    string
		genericFunc func(msg *model.Message)
	}{
		{
			operation:   constants.OptUpdate,
			resource:    msg.Router.Resource,
			genericFunc: am.proxyManger.UpdateAlarm,
		},
		{
			operation:   constants.OptQuery,
			resource:    constants.QueryAllAlarm,
			genericFunc: am.proxyManger.QueryAllAlarm,
		},
	}
	msgRegKey := msg.Router.Option + ":" + msg.Router.Resource
	for _, reg := range registerInfoList {
		oldRegKey := reg.operation + ":" + reg.resource
		if oldRegKey == msgRegKey {
			reg.genericFunc(msg)
			return
		}
	}
	hwlog.RunLog.Errorf("unsupported msg, operation=%s, resource=%s",
		msg.Router.Option, msg.Router.Resource)
}
