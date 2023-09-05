// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package alarmmanager for alarm-manager module init
package alarmmanager

import (
	"context"
	"errors"
	"net/http"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"alarm-manager/pkg/monitors"
	"alarm-manager/pkg/utils"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
)

type handlerFunc func(req interface{}) (interface{}, error)

type alarmManager struct {
	dbPath string
	enable bool
	ctx    context.Context
}

// NewAlarmManager create cert manager
func NewAlarmManager(dbPath string, enable bool, ctx context.Context) model.Module {
	cm := &alarmManager{
		dbPath: dbPath,
		enable: enable,
		ctx:    ctx,
	}
	return cm
}

func (am *alarmManager) Name() string {
	return utils.AlarmModuleName
}

func (am *alarmManager) Enable() bool {
	return am.enable
}

func methodSelect(req *model.Message) (interface{}, error) {
	method, exit := handlerFuncMap[common.Combine(req.GetOption(), req.GetResource())]
	if !exit {
		hwlog.RunLog.Errorf("handler func is not exist, option: %s, resource: %s", req.GetOption(),
			req.GetResource())
		return nil, errors.New("handler func is not exist")
	}
	return method(req.GetContent())
}

func (am *alarmManager) Start() {
	go am.startMonitoring()
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
	msg, err := methodSelect(req)
	if err != nil {
		hwlog.RunLog.Errorf("%s get method by option and resource failed: %s", am.Name(), err.Error())
		return
	}
	if msg == nil {
		return
	}

	resp, err := req.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("%s new response failed: %s", am.Name(), err.Error())
		return
	}

	resp.FillContent(msg)
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
	common.Combine(http.MethodPost, requests.ReportAlarmRouter):     dealAlarmReq,
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
