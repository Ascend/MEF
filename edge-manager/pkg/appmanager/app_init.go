// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager
package appmanager

import (
	"context"
	"fmt"

	"edge-manager/pkg/database"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

type handlerFunc func(req interface{}) common.RespMsg

type appManager struct {
	enable bool
	ctx    context.Context
}

// NewAppManager create app manager
func NewAppManager(enable bool) *appManager {
	am := &appManager{
		enable: enable,
		ctx:    context.Background(),
	}
	return am
}

func (app *appManager) Name() string {
	return common.AppManagerName
}

func (app *appManager) Enable() bool {
	if app.enable {
		if err := initAppTable(); err != nil {
			hwlog.RunLog.Errorf("module (%s) init database table failed, cannot enable", common.AppManagerName)
			return !app.enable
		}
	}
	return app.enable
}

func (app *appManager) Start() {
	for {
		select {
		case _, ok := <-app.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("has listened stop signal")
			return
		default:
		}
		req, err := modulemanager.ReceiveMessage(common.AppManagerName)
		hwlog.RunLog.Debugf("%s receive requst from restful service", common.AppManagerName)
		if err != nil {
			hwlog.RunLog.Errorf("%s receive requst from restful service failed", common.AppManagerName)
			continue
		}
		msg := methodSelect(req)
		if msg == nil {
			hwlog.RunLog.Errorf("%s get method by option and resource failed", common.AppManagerName)
			continue
		}
		resp, err := req.NewResponse()
		if err != nil {
			hwlog.RunLog.Errorf("%s new response failed", common.AppManagerName)
			continue
		}
		resp.FillContent(msg)
		if err = modulemanager.SendMessage(resp); err != nil {
			hwlog.RunLog.Errorf("%s send response failed", common.AppManagerName)
			continue
		}
	}
}

func methodSelect(req *model.Message) *common.RespMsg {
	var res common.RespMsg
	method, exit := appMethodList()[combine(req.GetOption(), req.GetResource())]
	if !exit {
		return nil
	}
	res = method(req.GetContent())
	return &res
}

func initAppTable() error {
	if err := database.CreateTableIfNotExists(AppInfo{}); err != nil {
		hwlog.RunLog.Error("create app information database table failed")
		return err
	}
	if err := database.CreateTableIfNotExists(AppContainer{}); err != nil {
		hwlog.RunLog.Error("create app container database table failed")
		return err
	}
	return nil
}

func appMethodList() map[string]handlerFunc {
	return map[string]handlerFunc{
		combine(common.Create, common.App): CreateApp,
		combine(common.List, common.App):   ListAppInfo,
		combine(common.Deploy, common.App): DeployApp,
	}
}

func combine(option, resource string) string {
	return fmt.Sprintf("%s%s", option, resource)
}
