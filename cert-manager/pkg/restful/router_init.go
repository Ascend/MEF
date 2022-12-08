// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restful init restful service
package restful

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

var (
	// BuildNameStr the program name
	BuildNameStr string
	// BuildVersionStr the program version
	BuildVersionStr string
)

// Service cert manager service init
type Service struct {
	enable bool
	engine *gin.Engine
	port   int
	ip     string
}

func initGin() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	return gin.New()
}

// NewRestfulService new restful service
func NewRestfulService(enable bool, ip string, port int) *Service {
	nm := &Service{
		enable: enable,
		engine: initGin(),
		port:   port,
		ip:     ip,
	}
	return nm
}

// Name for RestfulService name
func (r *Service) Name() string {
	return common.CertManagerService
}

// Start for RestfulService start
func (r *Service) Start() {
	r.engine.Use(common.LoggerAdapter())
	setRouter(r.engine)
	hwlog.RunLog.Info("start cert manager http server now...")
	err := r.engine.Run(fmt.Sprintf(":%d", r.port))
	if err != nil {
		hwlog.RunLog.Errorf("start restful at %d fail", r.port)
	}
}

// Enable for RestfulService enable
func (r *Service) Enable() bool {
	return r.enable
}
