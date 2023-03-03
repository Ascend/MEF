// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package usermgr this package is for manage user
package usermgr

import (
	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common/certutils"
	"huawei.com/mindxedge/base/common/httpsmgr"
	"huawei.com/mindxedge/base/common/restfulmgr"
	"nginx-manager/pkg/nginxcom"
)

// UserRestfulService [struct] for User Restful Service
type UserRestfulService struct {
	enable   bool
	httpsSvr *httpsmgr.HttpsServer
}

// NewRestfulService new restful service
func NewRestfulService(enable bool, ip string, port int) *UserRestfulService {
	nm := &UserRestfulService{
		enable: enable,
		httpsSvr: &httpsmgr.HttpsServer{
			IP:   ip,
			Port: port,
			TlsCertPath: certutils.TlsCertInfo{
				RootCaPath: nginxcom.RootCaPath,
				CertPath:   nginxcom.UserCertFile,
				KeyPath:    nginxcom.UserCertKeyFile,
				SvrFlag:    true,
				KmcCfg:     nil,
			},
		},
	}
	return nm
}

// Start for RestfulService start
func (r *UserRestfulService) Start() {
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

// Name for UserRestfulService name
func (r *UserRestfulService) Name() string {
	return nginxcom.UserRestfulServiceName
}

// Enable for UserRestfulService enable
func (r *UserRestfulService) Enable() bool {
	return r.enable
}

func setRouter(engine *gin.Engine) {
	restfulmgr.InitRouter(engine, routers)
}
