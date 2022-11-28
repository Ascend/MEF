// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/util"
)

func setRouter(engine *gin.Engine) {
	engine.Use(gin.Recovery())
	nodeRouter(engine)
	appRouter(engine)
	wsRouter(engine)
	templateRouter(engine)
}

func nodeRouter(engine *gin.Engine) {
	node := engine.Group("/edgemanager/v1/node")
	{
		node.POST("/", createEdgeNode)
		node.GET("/:id", getNodeDetail)
		node.PUT("/", modifyNode)
		node.GET("/num", getNodeStatistics)
		node.GET("/list/managed", listNodeManaged)
		node.GET("/list/unmanaged", listNodeUnManaged)
	}
	nodeGroup := engine.Group("/edgemanager/v1/nodegroup")
	{
		nodeGroup.POST("/", createEdgeNodeGroup)
		nodeGroup.GET("/", listEdgeNodeGroup)
		nodeGroup.GET("/:id", getEdgeNodeGroupDetail)
	}
}

func appRouter(engine *gin.Engine) {
	app := engine.Group("/edgemanager/v1/app")
	{
		app.POST("/", createApp)
		app.GET("/list", listAppsInfo)
		app.POST("/deploy", deployApp)
		app.DELETE("/", deleteApp)
	}
}

func wsRouter(engine *gin.Engine) {
	v1 := engine.Group("/edgemanager/v1/ws")
	{
		v1.POST("/", upgradeSfw)
	}
}

func templateRouter(engine *gin.Engine) {
	v1 := engine.Group("/edgemanager/v1/apptemplate")
	{
		v1.POST("/", createTemplate)
		v1.POST("/delete", deleteTemplate)
		v1.PUT("/", modifyTemplate)
		v1.GET("/", getTemplateDetail)
		v1.GET("/list", getTemplates)
	}
}

func pageUtil(c *gin.Context) (util.ListReq, error) {
	input := util.ListReq{}
	var err error
	// for slice page on ucd
	input.PageNum, err = strconv.ParseUint(c.Query("pageNum"), common.BaseHex, common.BitSize64)
	if err != nil {
		return input, err
	}
	input.PageSize, err = strconv.ParseUint(c.Query("pageSize"), common.BaseHex, common.BitSize64)
	if err != nil {
		return input, err
	}
	// for fuzzy query
	input.Name = c.Query("name")
	return input, nil
}
