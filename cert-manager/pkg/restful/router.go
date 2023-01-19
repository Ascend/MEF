// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restful this file is for setup router
package restful

import (
	"github.com/gin-gonic/gin"
)

func setRouter(engine *gin.Engine) {
	v1 := engine.Group("certmanager/v1/certificates")
	{
		v1.POST("/import", importRootCa)
		v1.GET("/rootca", queryRootCA)
		v1.POST("/service", issueServiceCa)
		v1.GET("/alert", queryAlert)
	}
	engine.GET("/certmanager/v1/version", versionQuery)
}
