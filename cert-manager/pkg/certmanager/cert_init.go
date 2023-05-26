// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package certmanager cert manager module
package certmanager

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

type certManager struct {
	enable bool
	ctx    context.Context
}

// NewCertManager create cert manager
func NewCertManager(enable bool) model.Module {
	cm := &certManager{
		enable: enable,
		ctx:    context.Background(),
	}
	return cm
}

func (cm *certManager) Name() string {
	return common.CertManagerName
}

func (cm *certManager) Enable() bool {
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

func (cm *certManager) Start() {
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

func (cm *certManager) dispatch(req *model.Message) {
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
	certUrlRootPath      = "/certmanager/v1/certificates"
	innerCertUrlRootPath = "/inner/v1/certificates"
)

var handlerFuncMap = map[string]handlerFunc{
	common.Combine(http.MethodPost, filepath.Join(certUrlRootPath, "import")):      importRootCa,
	common.Combine(http.MethodPost, filepath.Join(certUrlRootPath, "delete-cert")): deleteRootCa,

	common.Combine(http.MethodGet, filepath.Join(innerCertUrlRootPath, "rootca")):   queryRootCa,
	common.Combine(http.MethodPost, filepath.Join(innerCertUrlRootPath, "service")): issueServiceCa,
}
