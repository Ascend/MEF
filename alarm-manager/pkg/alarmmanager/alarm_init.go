// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package alarmmanager for alarm-manager module init
package alarmmanager

import (
	"context"
	"net/http"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"
)

type handlerFunc func(req interface{}) common.RespMsg

type alarmManager struct {
	enable bool
	ctx    context.Context
}

// NewAlarmManager create cert manager
func NewAlarmManager(enable bool, ctx context.Context) model.Module {
	cm := &alarmManager{
		enable: enable,
		ctx:    ctx,
	}
	return cm
}

func (cm *alarmManager) Name() string {
	return AlarmModuleName
}

func (cm *alarmManager) Enable() bool {
	return cm.enable
}

func methodSelect(req *model.Message) *common.RespMsg {
	var res common.RespMsg
	method, exit := handlerFuncMap[common.Combine(req.GetOption(), req.GetResource())]
	if !exit {
		hwlog.RunLog.Errorf("handler func is not exist, option: %s, resource: %s", req.GetOption(),
			req.GetResource())
		return nil
	}
	res = method(req.GetContent())
	return &res
}

func (cm *alarmManager) Start() {
	for {
		select {
		case _, ok := <-cm.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("has listened stop signal")
			return
		default:
		}

		req, err := modulemgr.ReceiveMessage(cm.Name())
		hwlog.RunLog.Infof("%s receive request from restful service", cm.Name())
		if err != nil {
			hwlog.RunLog.Errorf("%s receive request from restful service failed", cm.Name())
			continue
		}
		go cm.dispatch(req)
	}
}

func (cm *alarmManager) dispatch(req *model.Message) {
	msg := methodSelect(req)
	if msg == nil {
		hwlog.RunLog.Errorf("%s get method by option and resource failed", cm.Name())
		return
	}
	resp, err := req.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("%s new response failed", cm.Name())
		return
	}
	resp.FillContent(msg)
	if err = modulemgr.SendMessage(resp); err != nil {
		hwlog.RunLog.Errorf("%s send response failed", cm.Name())
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
	common.Combine(http.MethodGet, listAlarmRouter):      listAlarms,
	common.Combine(http.MethodGet, getAlarmDetailRouter): getAlarmDetail,
	common.Combine(http.MethodGet, listEventsRouter):     listEvents,
	common.Combine(http.MethodGet, getEventDetailRouter): getEventDetail,
}
