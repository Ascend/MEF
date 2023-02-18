// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restful init restful service
package restful

import (
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/certutils"
	"huawei.com/mindxedge/base/common/httpsmgr"

	"cert-manager/pkg/certconstant"
	"cert-manager/pkg/certmanager"
)

var (
	// BuildNameStr the program name
	BuildNameStr string
	// BuildVersionStr the program version
	BuildVersionStr string
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
				RootCaPath:    certconstant.RootCaPath,
				CertPath:      certconstant.ServerCertPath,
				KeyPath:       certconstant.ServerKeyPath,
				SvrFlag:       true,
				IgnoreCltCert: false,
				KmcCfg:        nil,
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
	go func() {
		err := certmanager.CheckAndCreateRootCa()
		if err != nil {
			hwlog.RunLog.Errorf("start cert restful at %d failed, check cert failed: %v", r.httpsSvr.Port, err)
			return
		}
	}()

	err := r.httpsSvr.Init()
	if err != nil {
		hwlog.RunLog.Errorf("start restful at %d failed, init https server failed: %v", r.httpsSvr.Port, err)
		return
	}
	hwlog.RunLog.Info("init cert manager https server success")
	err = r.httpsSvr.RegisterRoutes(setRouter)
	if err != nil {
		hwlog.RunLog.Errorf("start restful at %d failed, set routers failed: %v", r.httpsSvr.Port, err)
		return
	}
	hwlog.RunLog.Info("set cert manager https routers success")
	hwlog.RunLog.Info("start cert manager https server ......")
	err = r.httpsSvr.Start()
	if err != nil {
		hwlog.RunLog.Errorf("start restful at %d failed, listen failed: %v", r.httpsSvr.Port, err)
	}
}

// Enable for RestfulService enable
func (r *Service) Enable() bool {
	return r.enable
}
