// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package alarmmanager for alarm-manager module init
package alarmmanager

import (
	"context"
	"net/http"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"alarm-manager/pkg/monitors"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
)

type handlerFunc func(msg *model.Message) interface{}

type alarmManager struct {
	dbPath string
	enable bool
	ctx    context.Context
}

// NewAlarmManager create cert manager
func NewAlarmManager(dbPath string, enable bool, ctx context.Context) model.Module {
	return &alarmManager{
		dbPath: dbPath,
		enable: enable,
		ctx:    ctx,
	}
}

func (am *alarmManager) Name() string {
	return common.AlarmManagerName
}

func (am *alarmManager) Enable() bool {
	return am.enable
}

func methodSelect(req *model.Message) interface{} {
	method, exit := handlerFuncMap[common.Combine(req.GetOption(), req.GetResource())]
	if !exit {
		hwlog.RunLog.Errorf("handler func is not exist, option: %s, resource: %s", req.GetOption(),
			req.GetResource())
		return nil
	}
	return method(req)
}

func (am *alarmManager) Start() {
	go am.startMonitoring()
	go am.checkAlarmNum()
	for {
		select {
		case _, ok := <-am.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("has listened stop signal")
			return
		default:
		}

		req, err := modulemgr.ReceiveMessage(am.Name())
		if err != nil {
			hwlog.RunLog.Errorf("%s receive request from restful service failed", am.Name())
			continue
		}

		go am.dispatch(req)
	}
}

func (am *alarmManager) dispatch(req *model.Message) {
	msg := methodSelect(req)
	if msg == nil {
		return
	}

	resp, err := req.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("%s new response failed: %s", am.Name(), err.Error())
		return
	}

	if err = resp.FillContent(msg); err != nil {
		hwlog.RunLog.Errorf("%s fill content into resp failed: %s", am.Name(), err.Error())
		return
	}
	if err = modulemgr.SendMessage(resp); err != nil {
		hwlog.RunLog.Errorf("%s send response failed: %s", am.Name(), err.Error())
		return
	}
}

var (
	listAlarmRouter      = "/alarmmanager/v1/alarms"
	listEventsRouter     = "/alarmmanager/v1/events"
	getAlarmDetailRouter = "/alarmmanager/v1/alarm"
	getEventDetailRouter = "/alarmmanager/v1/event"
)

var handlerFuncMap = map[string]handlerFunc{
	common.Combine(http.MethodGet, listAlarmRouter):                 listAlarms,
	common.Combine(http.MethodGet, getAlarmDetailRouter):            getAlarmDetail,
	common.Combine(http.MethodGet, listEventsRouter):                listEvents,
	common.Combine(http.MethodGet, getEventDetailRouter):            getEventDetail,
	common.Combine(http.MethodPost, requests.ReportAlarmRouter):     dealAlarmsReq,
	common.Combine(common.Delete, requests.ClearOneNodeAlarmRouter): dealNodeClearReq,
}

func (am *alarmManager) startMonitoring() {
	var alarmMangerList []monitors.AlarmMonitor
	alarmMangerList = monitors.GetAlarmMonitorList(am.dbPath)
	for _, alarm := range alarmMangerList {
		if alarm != nil {
			go alarm.Monitoring(am.ctx)
		}
	}
}

func (am *alarmManager) checkAlarmNum() {
	const checkInterval = 5 * time.Minute
	tick := time.NewTicker(checkInterval)
	defer tick.Stop()
	for {
		select {
		case <-am.ctx.Done():
			hwlog.RunLog.Info("catch stop signal, channel is closed")
			return
		case <-tick.C:
			if err := clearEdgeAlarms(); err != nil {
				continue
			}
		}
	}
}

func clearEdgeAlarms() error {
	total, err := common.GetItemCount(AlarmInfo{})
	if err != nil {
		hwlog.RunLog.Errorf("get number of table alarm info failed, error: %v", err)
		return err
	}

	const allowMaxAlarm = 100000
	if total >= allowMaxAlarm {
		// An exception occurs when the number of alarms reaches the upper limit. Record error logs directly.
		hwlog.RunLog.Error("number of table alarm info is enough, need to be cleared")
		if err = AlarmDbInstance().DeleteEdgeAlarm(); err != nil {
			hwlog.RunLog.Errorf("clear all alarms from edge failed: %s", err.Error())
			return err
		}
		hwlog.RunLog.Error("clear all alarms from edge success")
	}
	return nil
}
