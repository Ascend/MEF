// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package apptemplatemanager to  provide containerized application template management.
package apptemplatemanager

import (
	"context"
	"edge-manager/pkg/database"
	"errors"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

type handlerFunc func(req interface{}) common.RespMsg

// ModuleAppTemplateManager module app template manager
type ModuleAppTemplateManager struct {
	routes map[string]handlerFunc
	ctx    context.Context
	cancel context.CancelFunc
	enable bool
}

// NewTemplateManager new node manager
func NewTemplateManager(enable bool) *ModuleAppTemplateManager {
	module := &ModuleAppTemplateManager{
		enable: enable,
		ctx:    context.Background(),
		routes: make(map[string]handlerFunc),
	}
	module.registerRoutes()
	module.ctx, module.cancel = context.WithCancel(context.Background())
	return module
}

// Name module name
func (manager *ModuleAppTemplateManager) Name() string {
	return common.TemplateManagerName
}

// Enable module enable
func (manager *ModuleAppTemplateManager) Enable() bool {
	if manager.enable {
		if err := initTable(); err != nil {
			hwlog.RunLog.Errorf("module %v init database table failed,error:%v", common.TemplateManagerName, err)
			return !manager.enable
		}
	}
	return manager.enable
}

// Stop module stop
func (manager *ModuleAppTemplateManager) Stop() bool {
	manager.cancel()
	return true
}

// Start module start running
func (manager *ModuleAppTemplateManager) Start() {
	for {
		select {
		case _, ok := <-manager.ctx.Done():
			if !ok {
				hwlog.RunLog.Infof("catch stop signal channel is closed")
			}
			hwlog.RunLog.Infof("has listened stop signal")
			return
		default:
		}
		receivedMsg, err := modulemanager.ReceiveMessage(manager.Name())
		if err != nil {
			hwlog.RunLog.Errorf("get receive module message failed,error:%v", err)
			continue
		}
		param := receivedMsg.GetContent()
		result := manager.handle(receivedMsg.GetResource(), receivedMsg.GetOption(), param)
		var respMsg *model.Message
		respMsg, err = receivedMsg.NewResponse()
		if err != nil {
			hwlog.RunLog.Errorf("new response module message failed,error:%v", err)
			continue
		}
		respMsg.FillContent(result)
		if err = modulemanager.SendMessage(respMsg); err != nil {
			hwlog.RunLog.Errorf("send response module message failed,error:%v", err)
		}
	}
}

func initTable() error {
	if err := database.CreateTableIfNotExists(AppTemplateDb{}); err != nil {
		return errors.New("create app template group database table failed")
	}
	if err := database.CreateTableIfNotExists(TemplateContainerDb{}); err != nil {
		return errors.New("create app template container database table failed")
	}
	return nil
}

func (manager *ModuleAppTemplateManager) handle(resource, option string, req interface{}) common.RespMsg {
	name := routeName(resource, option)
	if _, ok := manager.routes[name]; !ok {
		hwlog.RunLog.Errorf("module resource (%v) or option (%v) not found", resource, option)
		return common.RespMsg{Status: common.ErrorResourceOptionNotFound}
	}
	return manager.routes[name](req)
}

func (manager *ModuleAppTemplateManager) registerRoutes() {
	manager.setRoute(common.AppTemplate, common.Create, CreateTemplate)
	manager.setRoute(common.AppTemplate, common.Update, ModifyTemplate)
	manager.setRoute(common.AppTemplate, common.Delete, DeleteTemplate)
	manager.setRoute(common.AppTemplate, common.List, GetTemplates)
	manager.setRoute(common.AppTemplate, common.Get, GetTemplateDetail)
}

func (manager *ModuleAppTemplateManager) setRoute(resource, option string, handler handlerFunc) {
	manager.routes[routeName(resource, option)] = handler
}

func routeName(resource, option string) string {
	return resource + option
}
