// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager
package appmanager

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
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
	method, exit := handlerFuncMap[combine(req.GetOption(), req.GetResource())]
	if !exit {
		hwlog.RunLog.Errorf("handler func is not exist, option: %s, resource: %s", req.GetOption(),
			req.GetResource())
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
	if err := database.CreateTableIfNotExists(AppDaemonSet{}); err != nil {
		hwlog.RunLog.Error("create app daemon set database table failed")
		return err
	}
	if err := database.CreateTableIfNotExists(AppTemplateDb{}); err != nil {
		hwlog.RunLog.Error("create app template instance database table failed")
		return err
	}

	return nil
}

var (
	appUrlRootPath      = "/edgemanager/v1/app"
	appTemplateRootPath = "/edgemanager/v1/apptemplate"
)

var handlerFuncMap = map[string]handlerFunc{
	combine(http.MethodPost, appUrlRootPath):                                createApp,
	combine(http.MethodGet, appUrlRootPath):                                 queryApp,
	combine(http.MethodPatch, appUrlRootPath):                               updateApp,
	combine(http.MethodGet, filepath.Join(appUrlRootPath, "list")):          listAppInfo,
	combine(http.MethodPost, filepath.Join(appUrlRootPath, "deployment")):   deployApp,
	combine(http.MethodDelete, filepath.Join(appUrlRootPath, "deployment")): unDeployApp,
	combine(http.MethodGet, filepath.Join(appUrlRootPath, "deployment")):    listAppInstances,
	combine(http.MethodDelete, appUrlRootPath):                              deleteApp,
	combine(http.MethodGet, filepath.Join(appUrlRootPath, "node")):          listAppInstancesByNode,

	combine(http.MethodPost, appTemplateRootPath):                       createTemplate,
	combine(http.MethodPatch, appTemplateRootPath):                      updateTemplate,
	combine(http.MethodDelete, appTemplateRootPath):                     deleteTemplate,
	combine(http.MethodGet, appTemplateRootPath):                        getTemplate,
	combine(http.MethodGet, filepath.Join(appTemplateRootPath, "list")): getTemplates,
	combine(common.Get, common.AppInstanceByNodeGroup):                  getAppInstanceCountByNodeGroup,
}

func combine(option, resource string) string {
	return fmt.Sprintf("%s%s", option, resource)
}
