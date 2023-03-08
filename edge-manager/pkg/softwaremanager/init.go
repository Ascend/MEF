// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package softwaremanager to manage software
package softwaremanager

import (
	"context"
	"net/http"
	"path/filepath"

	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/database"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

type handlerFunc func(req interface{}) common.RespMsg

// SfwManager [struct] to describe software manager
type SfwManager struct {
	enable bool
	ctx    context.Context
}

// NewSftManager create config manager
func NewSftManager(enable bool) model.Module {
	im := &SfwManager{
		enable: enable,
		ctx:    context.Background(),
	}
	return im
}

// Name [method] module name
func (sm *SfwManager) Name() string {
	return common.SoftwareManagerName
}

// Enable [method] for SfwManager enable
func (sm *SfwManager) Enable() bool {
	if sm.enable {
		if err := database.CreateTableIfNotExists(SoftwareInfo{}); err != nil {
			hwlog.RunLog.Error("create sft info database table failed")
			return !sm.enable
		}
	}
	return sm.enable
}

// Start [method] for SfwManager start
func (sm *SfwManager) Start() {
	for {
		select {
		case _, ok := <-sm.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("has listened stop signal")
			return
		default:
		}

		req, err := modulemanager.ReceiveMessage(sm.Name())
		if err != nil {
			hwlog.RunLog.Errorf("%s receive message failed: %v", sm.Name(), err)
			continue
		}

		hwlog.RunLog.Infof("module [%s] receive message option:%s, resource:%s", sm.Name(),
			req.GetOption(), req.GetResource())

		go sm.dispatch(req)
	}
}

func (sm *SfwManager) dispatch(req *model.Message) {
	msg := methodSelect(req)
	if msg == nil {
		hwlog.RunLog.Errorf("%s get method by option and resource failed", sm.Name())
		return
	}

	if !req.GetIsSync() {
		return
	}

	resp, err := req.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("%s new response failed", sm.Name())
		return
	}
	resp.FillContent(msg)
	if err = modulemanager.SendMessage(resp); err != nil {
		hwlog.RunLog.Errorf("%s send response failed", sm.Name())
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

var (
	configUrlRootPath = "/edgemanager/v1/software"
)

var handlerFuncMap = map[string]handlerFunc{
	common.Combine(http.MethodPost, filepath.Join(configUrlRootPath, "auth-info")): updateAuthInfo,
	common.Combine(http.MethodPost, filepath.Join(configUrlRootPath, "url-info")):  updateSftUrlInfo,
	common.Combine(common.Inner, common.ResSfwDownloadInfo):                        innerGetSftDownloadInfo,
}
