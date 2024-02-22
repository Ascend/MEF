// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restful init restful service
package restful

import (
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509/certutils"

	"alarm-manager/pkg/utils"

	"huawei.com/mindxedge/base/common"
)

// Service cert manager service init
type Service struct {
	enable   bool
	httpsSvr *httpsmgr.HttpsServer
}

// NewRestfulService new restful service
func NewRestfulService(enable bool, conf *httpsmgr.HttpsServer) *Service {
	conf.TlsCertPath = certutils.TlsCertInfo{
		RootCaPath: utils.RootCaPath,
		CertPath:   utils.ServerCertPath,
		KeyPath:    utils.ServerKeyPath,
		SvrFlag:    true,
		KmcCfg:     nil,
		WithBackup: true,
	}

	return &Service{
		enable:   enable,
		httpsSvr: conf,
	}
}

// Name for RestfulService name
func (r *Service) Name() string {
	return common.RestfulServiceName
}

// Start for RestfulService start
func (r *Service) Start() {
	if err := r.httpsSvr.Init(); err != nil {
		hwlog.RunLog.Errorf("start restful at %d failed, init https server failed: %v", r.httpsSvr.Port, err)
		return
	}
	hwlog.RunLog.Info("init alarm manager https server success")
	if err := r.httpsSvr.RegisterRoutes(setRouter); err != nil {
		hwlog.RunLog.Errorf("start restful at %d failed, set routers failed: %v", r.httpsSvr.Port, err)
		return
	}
	hwlog.RunLog.Info("set alarm manager https routers success")
	hwlog.RunLog.Info("start alarm manager https server ......")
	if err := r.httpsSvr.Start(); err != nil {
		hwlog.RunLog.Errorf("start restful at %d failed, listen failed: %v", r.httpsSvr.Port, err)
	}
}

// Enable for RestfulService enable
func (r *Service) Enable() bool {
	return r.enable
}
