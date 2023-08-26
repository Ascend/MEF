// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restful init restful service
package restful

import (
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509/certutils"

	"alarm-manager/pkg/alarmmanager"
	"huawei.com/mindxedge/base/common"
)

// Service cert manager service init
type Service struct {
	enable   bool
	httpsSvr *httpsmgr.HttpsServer
}

// NewRestfulService new restful service
func NewRestfulService(enable bool, ip string, port int) *Service {
	nm := &Service{
		enable: enable,
		httpsSvr: &httpsmgr.HttpsServer{
			IP:   ip,
			Port: port,
			TlsCertPath: certutils.TlsCertInfo{
				RootCaPath: alarmmanager.RootCaPath,
				CertPath:   alarmmanager.ServerCertPath,
				KeyPath:    alarmmanager.ServerKeyPath,
				SvrFlag:    true,
				KmcCfg:     nil,
			},
		},
	}
	return nm
}

// Name for RestfulService name
func (r *Service) Name() string {
	return common.RestfulServiceName
}

// Start for RestfulService start
func (r *Service) Start() {
	err := r.httpsSvr.Init()
	if err != nil {
		hwlog.RunLog.Errorf("start restful at %d failed, init https server failed: %v", r.httpsSvr.Port, err)
		return
	}
	hwlog.RunLog.Info("init alarm manager https server success")
	err = r.httpsSvr.RegisterRoutes(setRouter)
	if err != nil {
		hwlog.RunLog.Errorf("start restful at %d failed, set routers failed: %v", r.httpsSvr.Port, err)
		return
	}
	hwlog.RunLog.Info("set alarm manager https routers success")
	hwlog.RunLog.Info("start alarm manager https server ......")
	err = r.httpsSvr.Start()
	if err != nil {
		hwlog.RunLog.Errorf("start restful at %d failed, listen failed: %v", r.httpsSvr.Port, err)
	}
}

// Enable for RestfulService enable
func (r *Service) Enable() bool {
	return r.enable
}
