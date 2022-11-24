// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package softwaremanager this file is for modular stuff
package softwaremanager

import (
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

type handlerFunc func(req interface{}) common.RespMsg

type softwareManager struct {
	enable bool
	ctx    chan struct{}
}

// Name module name
func (sr *softwareManager) Name() string {
	return common.SoftwareRepositoryName
}

// Enable module enable
func (sr *softwareManager) Enable() bool {
	return sr.enable
}

// Start module start
func (sr *softwareManager) Start() {
	for {
		select {
		case <-sr.ctx:
			return
		default:
		}
		req, err := modulemanager.ReceiveMessage(common.SoftwareRepositoryName)
		hwlog.RunLog.Debugf("%s receive request from software manager", common.SoftwareRepositoryName)
		if err != nil {
			hwlog.RunLog.Errorf("%s receive request from software manager failed", common.SoftwareRepositoryName)
			continue
		}
		msg := methodSelect(req)
		if msg == nil {
			hwlog.RunLog.Errorf("%s get method by option and resource failed", common.SoftwareRepositoryName)
			continue
		}
		resp, err := req.NewResponse()
		if err != nil {
			hwlog.RunLog.Errorf("%s new response failed", common.SoftwareRepositoryName)
			continue
		}
		resp.FillContent(msg)
		if err = modulemanager.SendMessage(resp); err != nil {
			hwlog.RunLog.Errorf("%s send response failed", common.SoftwareRepositoryName)
			continue
		}
	}
}

// NewSoftwareManager create SoftwareRepository module
func NewSoftwareManager(enable bool) model.Module {
	return &softwareManager{
		enable: enable,
		ctx:    make(chan struct{}),
	}
}

func methodSelect(req *model.Message) *common.RespMsg {
	var res common.RespMsg
	method, exit := nodeMethodList()[combine(req.GetOption(), req.GetResource())]
	if !exit {
		return nil
	}
	res = method(req.GetContent())
	return &res
}

func nodeMethodList() map[string]handlerFunc {
	return map[string]handlerFunc{}
}

func combine(option, resource string) string {
	return fmt.Sprintf("%s%s", option, resource)
}
