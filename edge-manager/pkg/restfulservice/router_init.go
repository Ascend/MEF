// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509/certutils"

	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/constants"
)

// EdgeMgrService [struct] for Edge Manager Service
type EdgeMgrService struct {
	enable   bool
	httpsSvr *httpsmgr.HttpsServer
}

// NewRestfulService new restful service
func NewRestfulService(enable bool, ip string, port int) *EdgeMgrService {
	nm := &EdgeMgrService{
		enable: enable,
		httpsSvr: &httpsmgr.HttpsServer{
			IP:           ip,
			Port:         port,
			WriteTimeout: common.EdgeManagerRestfulWriteTimeout,
			TlsCertPath: certutils.TlsCertInfo{
				RootCaPath: constants.RootCaPath,
				CertPath:   constants.ServerCertPath,
				KeyPath:    constants.ServerKeyPath,
				SvrFlag:    true,
				KmcCfg:     nil,
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
