// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"edge-manager/pkg/util"

	"huawei.com/mindxedge/base/common"
)

func setRouter(engine *gin.Engine) {
	engine.Use(gin.Recovery())
	nodeRouter(engine)
	appRouter(engine)
	softwareRouter(engine)
	connInfoRouter(engine)
	templateRouter(engine)
}

func nodeRouter(engine *gin.Engine) {
	node := engine.Group("/edgemanager/v1/node")
	{
		node.POST("/", createEdgeNode)
		node.GET("/stats", getNodeStatistics)
		node.GET("/:id", getNodeDetail)
		node.PATCH("/", modifyNode)
		node.DELETE("/", deleteNode)
		node.GET("/list/managed", listNodeManaged)
		node.GET("/list/unmanaged", listNodeUnManaged)
		node.POST("/add", addUnManagedNode)
	}
	nodeGroup := engine.Group("/edgemanager/v1/nodegroup")
	{
		nodeGroup.POST("/", createEdgeNodeGroup)
		nodeGroup.GET("/", listEdgeNodeGroup)
		nodeGroup.PATCH("/", modifyNodeGroup)
		nodeGroup.GET("/stats", getNodeGroupStatistics)
		nodeGroup.GET("/:id", getEdgeNodeGroupDetail)
		nodeGroup.POST("/node", addNodeToGroup)
		nodeGroup.DELETE("/node", deleteNodeFromGroup)
		nodeGroup.DELETE("/pod", batchDeleteNodeRelation)
		nodeGroup.DELETE("/", batchDeleteNodeGroup)
	}
}

func appRouter(engine *gin.Engine) {
	app := engine.Group("/edgemanager/v1/app")
	{
		app.POST("/", createApp)
		app.GET("/", queryApp)
		app.PATCH("/", updateApp)
		app.GET("/list", listAppsInfo)
		app.POST("/deployment", deployApp)
		app.DELETE("/deployment", unDeployApp)
		app.DELETE("/", deleteApp)
		app.GET("/deployment", listAppInstance)
		app.GET("/node", listAppInstanceByNode)
	}
}

func softwareRouter(engine *gin.Engine) {
	v1 := engine.Group("/edgemanager/v1/software")
	{
		v1.POST("/upgrade", upgradeSoftware)
	}
}

func connInfoRouter(engine *gin.Engine) {
	v1 := engine.Group("/edgemanager/v1/conninfo")
	{
		v1.POST("/update", updateConnInfo)
	}
}

func templateRouter(engine *gin.Engine) {
	v1 := engine.Group("/edgemanager/v1/apptemplate")
	{
		v1.POST("/", createTemplate)
		v1.DELETE("/", deleteTemplate)
		v1.PATCH("/", updateTemplate)
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

func getReqId(c *gin.Context, idName string) (uint64, error) {
	value, err := strconv.ParseUint(c.Query(idName), common.BaseHex, common.BitSize64)
	if err != nil {
		return 0, fmt.Errorf("id is invalid")
	}

	return value, nil
}

func getReqNodeId(c *gin.Context) (int64, error) {
	value, err := strconv.ParseInt(c.Query("nodeID"), common.BaseHex, common.BitSize64)
	if err != nil {
		return 0, fmt.Errorf("app id is invalid")
	}

	return value, nil
}
