// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
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
