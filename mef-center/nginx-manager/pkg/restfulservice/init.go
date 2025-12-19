// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package restfulservice to init restful service
package restfulservice

import (
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509/certutils"

	"huawei.com/mindxedge/base/common"
	"nginx-manager/pkg/nginxcom"
)

// NgxMgrServer [struct] for Edge Manager Service
type NgxMgrServer struct {
	enable   bool
	httpsSvr *httpsmgr.HttpsServer
}

// NewNgxMgrServer new nginx restful service
// use nginx-manager.crt as both client side and server side cert
func NewNgxMgrServer(enable bool, ip string, port int) *NgxMgrServer {
	nm := &NgxMgrServer{
		enable: enable,
		httpsSvr: &httpsmgr.HttpsServer{
			IP:          ip,
			Port:        port,
			SwitchLimit: false,
			TlsCertPath: certutils.TlsCertInfo{
				RootCaPath: nginxcom.RootCaPath,
				CertPath:   nginxcom.ClientCertFile,
				KeyPath:    nginxcom.ClientCertKeyFile,
				SvrFlag:    true,
			},
		},
	}
	return nm
}

// Enable for NgxMgrServer enable
func (ngx *NgxMgrServer) Enable() bool {
	return ngx.enable
}

// Name for NgxMgrServer name
func (ngx *NgxMgrServer) Name() string {
	return common.RestfulServiceName
}

// Start for NgxMgrServer start
func (ngx *NgxMgrServer) Start() {
	if err := ngx.httpsSvr.Init(); err != nil {
		hwlog.RunLog.Errorf("init nginx https server failed: %v", err)
		return
	}
	if err := ngx.httpsSvr.RegisterRoutes(setRouter); err != nil {
		hwlog.RunLog.Errorf("register nginx server routers failed: %v", err)
		return
	}
	if err := ngx.httpsSvr.Start(); err != nil {
		hwlog.RunLog.Errorf("start nginx https server failed: %v", err)
		return
	}
	hwlog.RunLog.Info("start nginx inner RESTful https server success")
}
