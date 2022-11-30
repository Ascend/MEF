// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice this file is for init restful service
package restfulservice

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager/model"
)

type restfulService struct {
	enable bool
	engine *gin.Engine
	ip     string
	port   int
}

// Name module name
func (r *restfulService) Name() string {
	return common.RestfulServiceName
}

// Enable module enable
func (r *restfulService) Enable() bool {
	return r.enable
}

// Start module start
func (r *restfulService) Start() {
	r.engine.Use(common.LoggerAdapter())
	setRouter(r.engine)
	hwlog.RunLog.Info("start http server now...")
	if err := r.engine.Run(fmt.Sprintf(":%d", r.port)); err != nil {
		hwlog.RunLog.Errorf("start restful at %d fail", r.port)
	}
}

// NewRestfulService init restful service
func NewRestfulService(enable bool, ip string, port int) model.Module {
	gin.SetMode(gin.ReleaseMode)
	return &restfulService{
		enable: enable,
		engine: gin.New(),
		ip:     ip,
		port:   port,
	}
}
