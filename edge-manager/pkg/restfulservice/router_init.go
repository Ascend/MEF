// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"edge-manager/pkg/common"
	"fmt"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"
)

var (
	// BuildNameStr the program name
	BuildNameStr string
	// BuildVersionStr the program version
	BuildVersionStr string
)

type restfulService struct {
	enable bool
	engine *gin.Engine
}

// NewRestfulService new restful service
func NewRestfulService(enable bool, g *gin.Engine) *restfulService {
	nm := &restfulService{
		enable: enable,
		engine: g,
	}
	return nm
}

// Name for RestfulService name
func (r *restfulService) Name() string {
	return common.RestfulServiceName
}

// Start for RestfulService start
func (r *restfulService) Start() {
	r.engine.Use(common.LoggerAdapter())
	setRouter(r.engine)
	r.engine.GET("/edgemanager/v1/version", versionQuery)
	return
}

// Enable for RestfulService enable
func (r *restfulService) Enable() bool {
	return r.enable
}

func versionQuery(c *gin.Context) {
	msg := fmt.Sprintf("%s version: %s", BuildNameStr, BuildVersionStr)
	hwlog.OpLog.Infof("query edge manager version: %s successfully", msg)
	common.ConstructResp(c, common.Success, "", msg)
}
