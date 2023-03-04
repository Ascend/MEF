// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager
package appmanager

import (
	"context"
	"net/http"
	"path/filepath"
	"time"

	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/database"
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
func NewAppManager(enable bool) model.Module {
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

		go app.dispatch(req)
	}
}

func (app *appManager) housekeeper() {
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

func (app *appManager) dispatch(req *model.Message) {
	msg := methodSelect(req)
	if msg == nil {
		hwlog.RunLog.Errorf("%s get method by option and resource failed", common.AppManagerName)
		return
	}
	resp, err := req.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("%s new response failed", common.AppManagerName)
		return
	}
	resp.FillContent(msg)
	if err = modulemanager.SendMessage(resp); err != nil {
		hwlog.RunLog.Errorf("%s send response failed", common.AppManagerName)
		return
	}
}

func methodSelect(req *model.Message) *common.RespMsg {
	var res common.RespMsg
	method, ok := handlerFuncMap[common.Combine(req.GetOption(), req.GetResource())]
	if !ok {
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
	if err := database.CreateTableIfNotExists(ConfigmapInfo{}); err != nil {
		hwlog.RunLog.Error("create configmap instance database table failed")
		return err
	}

	return nil
}

var (
	appUrlRootPath = "/edgemanager/v1/app"
)

var handlerFuncMap = map[string]handlerFunc{
	common.Combine(http.MethodPost, appUrlRootPath):                                           createApp,
	common.Combine(http.MethodGet, appUrlRootPath):                                            queryApp,
	common.Combine(http.MethodPatch, appUrlRootPath):                                          updateApp,
	common.Combine(http.MethodGet, filepath.Join(appUrlRootPath, "list")):                     listAppInfo,
	common.Combine(http.MethodPost, filepath.Join(appUrlRootPath, "deployment")):              deployApp,
	common.Combine(http.MethodPost, filepath.Join(appUrlRootPath, "deployment/batch-delete")): unDeployApp,
	common.Combine(http.MethodGet, filepath.Join(appUrlRootPath, "deployment")):               listAppInstancesById,
	common.Combine(http.MethodPost, filepath.Join(appUrlRootPath, "batch-delete")):            deleteApp,
	common.Combine(http.MethodGet, filepath.Join(appUrlRootPath, "node")):                     listAppInstancesByNode,
	common.Combine(http.MethodGet, filepath.Join(appUrlRootPath, "deployment/list")):          listAppInstances,

	common.Combine(common.Get, common.AppInstanceByNodeGroup): getAppInstanceCountByNodeGroup,
}
