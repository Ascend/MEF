// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package configmanager to init config manager
package configmanager

import (
	"context"
	"net/http"
	"path/filepath"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"
)

type handlerFunc func(req interface{}) common.RespMsg

type configManager struct {
	enable bool
	ctx    context.Context
}

// NewConfigManager create config manager
func NewConfigManager(enable bool) model.Module {
	im := &configManager{
		enable: enable,
		ctx:    context.Background(),
	}
	return im
}

func (cm *configManager) Name() string {
	return common.ConfigManagerName
}

func (cm *configManager) Enable() bool {
	return cm.enable
}

func (cm *configManager) Start() {
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
		if err != nil {
			hwlog.RunLog.Errorf("%s receive request from restful service failed", cm.Name())
			continue
		}

		go cm.dispatch(req)
	}
}

func (cm *configManager) dispatch(req *model.Message) {
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

var (
	configUrlRootPath      = "/edgemanager/v1/image"
	innerConfigUrlRootPath = "/inner/v1/image"
)

var handlerFuncMap = map[string]handlerFunc{
	common.Combine(http.MethodPost, filepath.Join(configUrlRootPath, "config")):      downloadConfig,
	common.Combine(http.MethodPost, filepath.Join(innerConfigUrlRootPath, "update")): updateConfig,
}
