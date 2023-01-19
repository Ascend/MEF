// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"fmt"
	"net/http"
	"strconv"

	"edge-manager/pkg/util"

	"github.com/gin-gonic/gin"

	"huawei.com/mindxedge/base/common/restfulmgr"

	"huawei.com/mindxedge/base/common"
)

var appRouterDispatchers = map[string][]restfulmgr.DispatcherItf{
	"/edgemanager/v1/app": {
		restfulmgr.GenericDispatcher{
			Method:      http.MethodPost,
			Destination: common.AppManagerName},
		appQueryDispatcher{restfulmgr.GenericDispatcher{
			Method:      http.MethodGet,
			Destination: common.AppManagerName}},
		restfulmgr.GenericDispatcher{
			Method:      http.MethodPatch,
			Destination: common.AppManagerName},
		appListDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/list",
			Method:       http.MethodGet,
			Destination:  common.AppManagerName}},
		restfulmgr.GenericDispatcher{
			RelativePath: "/deployment",
			Method:       http.MethodPost,
			Destination:  common.AppManagerName},
		restfulmgr.GenericDispatcher{
			RelativePath: "/deployment",
			Method:       http.MethodDelete,
			Destination:  common.AppManagerName},
		restfulmgr.GenericDispatcher{
			Method:      http.MethodDelete,
			Destination: common.AppManagerName},
		appQueryDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/deployment",
			Method:       http.MethodGet,
			Destination:  common.AppManagerName}},
		appInstanceDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/node",
			Method:       http.MethodGet,
			Destination:  common.AppManagerName}},
		restfulmgr.GenericDispatcher{
			RelativePath: "/configmap",
			Method:       http.MethodPost,
			Destination:  common.AppManagerName},
		restfulmgr.GenericDispatcher{
			RelativePath: "/configmap/batch-delete",
			Method:       http.MethodPost,
			Destination:  common.AppManagerName},
		restfulmgr.GenericDispatcher{
			RelativePath: "/configmap",
			Method:       http.MethodPatch,
			Destination:  common.AppManagerName},
		cmQueryDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/configmap",
			Method:       http.MethodGet,
			Destination:  common.AppManagerName}},
		cmListDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/configmap/list",
			Method:       http.MethodGet,
			Destination:  common.AppManagerName}},
	},
}

var templateRouterDispatchers = map[string][]restfulmgr.DispatcherItf{
	"/edgemanager/v1/apptemplate": {
		restfulmgr.GenericDispatcher{
			Method:      http.MethodPost,
			Destination: common.AppManagerName},
		restfulmgr.GenericDispatcher{
			Method:      http.MethodDelete,
			Destination: common.AppManagerName},
		restfulmgr.GenericDispatcher{
			Method:      http.MethodPatch,
			Destination: common.AppManagerName},
		templateDetailDispatcher{restfulmgr.GenericDispatcher{
			Method:      http.MethodGet,
			Destination: common.AppManagerName}},
		listTemplateDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/list",
			Method:       http.MethodGet,
			Destination:  common.AppManagerName}},
	},
}

func setRouter(engine *gin.Engine) {
	engine.Use(gin.Recovery())
	nodeRouter(engine)
	restfulmgr.InitRouter(engine, appRouterDispatchers)
	restfulmgr.InitRouter(engine, templateRouterDispatchers)
	softwareRouter(engine)
	connInfoRouter(engine)
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

func getReqID(c *gin.Context, idName string) (uint64, error) {
	value, err := strconv.ParseUint(c.Query(idName), common.BaseHex, common.BitSize64)
	if err != nil {
		return 0, fmt.Errorf("id is invalid")
	}

	return value, nil
}

func getReqNodeID(c *gin.Context) (int64, error) {
	value, err := strconv.ParseInt(c.Query("nodeID"), common.BaseHex, common.BitSize64)
	if err != nil {
		return 0, fmt.Errorf("app id is invalid")
	}

	return value, nil
}

func getReqIntID(c *gin.Context, idName string) (int64, error) {
	value, err := strconv.ParseInt(c.Query(idName), common.BaseHex, common.BitSize64)
	if err != nil {
		return 0, fmt.Errorf("id name [%s] is invalid", idName)
	}

	return value, nil
}
