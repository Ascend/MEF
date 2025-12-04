// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package restful init restful service
package restful

import (
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509/certutils"

	"cert-manager/pkg/certmanager"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// Service cert manager service init
type Service struct {
	enable   bool
	httpsSvr *httpsmgr.HttpsServer
}

// NewRestfulService new restful service
func NewRestfulService(enable bool, conf *httpsmgr.HttpsServer) *Service {
	conf.TlsCertPath = certutils.TlsCertInfo{
		RootCaPath: util.RootCaPath,
		CertPath:   util.ServerCertPath,
		KeyPath:    util.ServerKeyPath,
		SvrFlag:    true,
		KmcCfg:     nil,
		WithBackup: true,
	}
	nm := &Service{
		enable:   enable,
		httpsSvr: conf,
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
		toGenCertsNames := []string{common.WsCltName, common.ThirdPartyCertName}
		for _, name := range toGenCertsNames {
			if err := certmanager.CreateCaIfNotExit(name); err != nil {
				hwlog.RunLog.Errorf("start cert restful at %d failed, check cert failed: %v", r.httpsSvr.Port, err)
				return
			}
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
