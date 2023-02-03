// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/util"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/certutils"
	"huawei.com/mindxedge/base/common/httpsmgr"
)

// EdgeMgrService [struct] for Edge Manager Service
type EdgeMgrService struct {
	enable   bool
	httpsSvr *httpsmgr.HttpsServer
	ip       string
}

// NewRestfulService new restful service
func NewRestfulService(enable bool, port int) *EdgeMgrService {
	nm := &EdgeMgrService{
		enable: enable,
		httpsSvr: &httpsmgr.HttpsServer{
			Port: port,
			TlsCertPath: certutils.TlsCertInfo{
				RootCaPath:    util.RootCaPath,
				CertPath:      util.ServerCertPath,
				KeyPath:       util.ServerKeyPath,
				SvrFlag:       true,
				IgnoreCltCert: true,
				KmcCfg:        nil,
			},
		},
	}
	return nm
}

// Name for EdgeMgrService name
func (r *EdgeMgrService) Name() string {
	return common.RestfulServiceName
}

// Start for EdgeMgrService start
func (r *EdgeMgrService) Start() {
	err := r.httpsSvr.Init()
	if err != nil {
		hwlog.RunLog.Errorf("start restful at %d failed, init https server failed: %v", r.httpsSvr.Port, err)
		return
	}
	err = r.httpsSvr.RegisterRoutes(setRouter)
	if err != nil {
		hwlog.RunLog.Errorf("start restful at %d failed, set routers failed: %v", r.httpsSvr.Port, err)
		return
	}

	hwlog.RunLog.Info("start http server now...")
	err = r.httpsSvr.Start()
	if err != nil {
		hwlog.RunLog.Errorf("start restful at %d fail", r.httpsSvr.Port)
	}
}

// Enable for EdgeMgrService enable
func (r *EdgeMgrService) Enable() bool {
	return r.enable
}
