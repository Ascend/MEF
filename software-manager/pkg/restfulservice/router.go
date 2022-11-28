// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice this file is for setup router
package restfulservice

import (
	"github.com/gin-gonic/gin"
)

func setRouter(engine *gin.Engine) {
	engine.Use(gin.Recovery())
	softwareRouter(engine)
}

func softwareRouter(engine *gin.Engine) {
	v1 := engine.Group("/software-manager/v1/softwaremanager")
	{
		v1.DELETE("/", deleteSoftware)
		v1.POST("/", uploadSoftware)
		v1.GET("/repository", checkRepository)
		v1.GET("/", downloadSoftware)
		v1.GET("/url", getURL)
	}
}
