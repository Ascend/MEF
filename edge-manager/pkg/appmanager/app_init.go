// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager
package appmanager

import (
	"context"
	"fmt"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"

	"edge-manager/pkg/database"
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
		if err := appStatusService.initAppStatusService(); err != nil {
			hwlog.RunLog.Errorf("module (%s) init app service failed, cannot enable", common.AppManagerName)
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
		go app.housekeeper()

		req, err := modulemanager.ReceiveMessage(common.AppManagerName)
		hwlog.RunLog.Infof("%s receive request from restful service", common.AppManagerName)
		if err != nil {
			hwlog.RunLog.Errorf("%s receive request from restful service failed", common.AppManagerName)
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

func (app appManager) housekeeper() {
	for {
		delay := time.NewTimer(houseKeepingInterval)
		select {
		case _, ok := <-app.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("has listened stop signal")
			if !delay.Stop() {
				<-delay.C
			}
			return
		case <-delay.C:
			appStatusService.deleteTerminatingPod()
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
	if err := database.CreateTableIfNotExists(AppInstance{}); err != nil {
		hwlog.RunLog.Error("create app instance database table failed")
		return err
	}

	if err := database.CreateTableIfNotExists(AppTemplateDb{}); err != nil {
		hwlog.RunLog.Error("create app template instance database table failed")
		return err
	}

	return nil
}

func appMethodList() map[string]handlerFunc {
	return map[string]handlerFunc{
		combine(common.Create, common.App):   createApp,
		combine(common.Query, common.App):    queryApp,
		combine(common.Update, common.App):   updateApp,
		combine(common.List, common.App):     listAppInfo,
		combine(common.Deploy, common.App):   deployApp,
		combine(common.Undeploy, common.App): unDeployApp,

		combine(common.Delete, common.App):             deleteApp,
		combine(common.List, common.AppInstance):       listAppInstances,
		combine(common.List, common.AppInstanceByNode): listAppInstancesByNode,

		combine(common.Create, common.AppTemplate): createTemplate,
		combine(common.Update, common.AppTemplate): updateTemplate,
		combine(common.Delete, common.AppTemplate): deleteTemplate,
		combine(common.List, common.AppTemplate):   getTemplates,
		combine(common.Get, common.AppTemplate):    getTemplate,
	}
}

func combine(option, resource string) string {
	return fmt.Sprintf("%s%s", option, resource)
}
