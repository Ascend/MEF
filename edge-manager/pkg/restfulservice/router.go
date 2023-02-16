// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/config"
	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/restfulmgr"
)

var appRouterDispatchers = map[string][]restfulmgr.DispatcherItf{
	"/edgemanager/v1/app": {
		restfulmgr.GenericDispatcher{
			Method:      http.MethodPost,
			Destination: common.AppManagerName},
		queryDispatcher{restfulmgr.GenericDispatcher{
			Method:      http.MethodGet,
			Destination: common.AppManagerName}, "appID", false},
		restfulmgr.GenericDispatcher{
			Method:      http.MethodPatch,
			Destination: common.AppManagerName},
		listDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/list",
			Method:       http.MethodGet,
			Destination:  common.AppManagerName}},
		restfulmgr.GenericDispatcher{
			RelativePath: "/deployment",
			Method:       http.MethodPost,
			Destination:  common.AppManagerName},
		restfulmgr.GenericDispatcher{
			RelativePath: "/deployment/batch-delete",
			Method:       http.MethodPost,
			Destination:  common.AppManagerName},
		restfulmgr.GenericDispatcher{
			RelativePath: "/batch-delete",
			Method:       http.MethodPost,
			Destination:  common.AppManagerName},
		queryDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/deployment",
			Method:       http.MethodGet,
			Destination:  common.AppManagerName}, "appID", false},
		queryDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/node",
			Method:       http.MethodGet,
			Destination:  common.AppManagerName}, "nodeID", false},
		listDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/deployment/list",
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
		queryDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/configmap",
			Method:       http.MethodGet,
			Destination:  common.AppManagerName}, "configmapID", false},
		listDispatcher{restfulmgr.GenericDispatcher{
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
			RelativePath: "/batch-delete",
			Method:       http.MethodPost,
			Destination:  common.AppManagerName},
		restfulmgr.GenericDispatcher{
			Method:      http.MethodPatch,
			Destination: common.AppManagerName},
		queryDispatcher{restfulmgr.GenericDispatcher{
			Method:      http.MethodGet,
			Destination: common.AppManagerName}, "id", false},
		listDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/list",
			Method:       http.MethodGet,
			Destination:  common.AppManagerName}},
	},
}

func setRouter(engine *gin.Engine) {
	engine.GET("/edgemanager/v1/version", versionQuery)
	restfulmgr.InitRouter(engine, nodeRouterDispatchers)
	restfulmgr.InitRouter(engine, nodeGroupRouterDispatchers)
	restfulmgr.InitRouter(engine, appRouterDispatchers)
	restfulmgr.InitRouter(engine, templateRouterDispatchers)
	restfulmgr.InitRouter(engine, softwareRouterDispatchers)
	connInfoRouter(engine)
}
func versionQuery(c *gin.Context) {
	msg := fmt.Sprintf("%s version: %s", config.BuildName, config.BuildVersion)
	hwlog.RunLog.Infof("query edge manager version: %s successfully", msg)
	common.ConstructResp(c, common.Success, "", msg)
}

var nodeRouterDispatchers = map[string][]restfulmgr.DispatcherItf{
	"/edgemanager/v1/node": {
		restfulmgr.GenericDispatcher{
			RelativePath: "/stats",
			Method:       http.MethodGet,
			Destination:  common.NodeManagerName},
		queryDispatcher{restfulmgr.GenericDispatcher{
			Method:      http.MethodGet,
			Destination: common.NodeManagerName}, "id", false},
		restfulmgr.GenericDispatcher{
			Method:      http.MethodPatch,
			Destination: common.NodeManagerName},
		restfulmgr.GenericDispatcher{
			RelativePath: "/batch-delete",
			Method:       http.MethodPost,
			Destination:  common.NodeManagerName},
		listDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/list/managed",
			Method:       http.MethodGet,
			Destination:  common.NodeManagerName}},
		listDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/list/unmanaged",
			Method:       http.MethodGet,
			Destination:  common.NodeManagerName}},
		listDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/list",
			Method:       http.MethodGet,
			Destination:  common.NodeManagerName}},
		restfulmgr.GenericDispatcher{
			RelativePath: "/add",
			Method:       http.MethodPost,
			Destination:  common.NodeManagerName},
	},
}

var nodeGroupRouterDispatchers = map[string][]restfulmgr.DispatcherItf{
	"/edgemanager/v1/nodegroup": {
		restfulmgr.GenericDispatcher{
			Method:      http.MethodPost,
			Destination: common.NodeManagerName},
		restfulmgr.GenericDispatcher{
			RelativePath: "/stats",
			Method:       http.MethodGet,
			Destination:  common.NodeManagerName},
		queryDispatcher{restfulmgr.GenericDispatcher{
			Method:      http.MethodGet,
			Destination: common.NodeManagerName}, "id", false},
		restfulmgr.GenericDispatcher{
			Method:      http.MethodPatch,
			Destination: common.NodeManagerName},
		restfulmgr.GenericDispatcher{
			RelativePath: "/batch-delete",
			Method:       http.MethodPost,
			Destination:  common.NodeManagerName},
		listDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/list",
			Method:       http.MethodGet,
			Destination:  common.NodeManagerName}},
		restfulmgr.GenericDispatcher{
			RelativePath: "/node",
			Method:       http.MethodPost,
			Destination:  common.NodeManagerName},
		restfulmgr.GenericDispatcher{
			RelativePath: "/node/batch-delete",
			Method:       http.MethodPost,
			Destination:  common.NodeManagerName},
		restfulmgr.GenericDispatcher{
			RelativePath: "/pod/batch-delete",
			Method:       http.MethodPost,
			Destination:  common.NodeManagerName},
	},
}

var softwareRouterDispatchers = map[string][]restfulmgr.DispatcherItf{
	"/edgemanager/v1/software/edge": {
		restfulmgr.GenericDispatcher{
			RelativePath: "/download",
			Method:       http.MethodPost,
			Destination:  common.NodeMsgManagerName},
		restfulmgr.GenericDispatcher{
			RelativePath: "/upgrade",
			Method:       http.MethodPost,
			Destination:  common.NodeMsgManagerName},
		queryDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/version-info",
			Method:       http.MethodGet,
			Destination:  common.NodeMsgManagerName}, "serialNumber", true},
		queryDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/download-result",
			Method:       http.MethodGet,
			Destination:  common.NodeMsgManagerName}, "serialNumber", true},
	},
}

func connInfoRouter(engine *gin.Engine) {
	v1 := engine.Group("/edgemanager/v1/conninfo")
	{
		v1.POST("/update", updateConnInfo)
	}
}

func pageUtil(c *gin.Context) (types.ListReq, error) {
	input := types.ListReq{}
	var err error
	// for slice page on ucd
	input.PageNum, err = getIntReqPara(c, "pageNum")
	if err != nil {
		return input, err
	}
	input.PageSize, err = getIntReqPara(c, "pageSize")
	if err != nil {
		return input, err
	}
	// for fuzzy query
	input.Name = c.Query("name")
	return input, nil
}

func getIntReqPara(c *gin.Context, paraName string) (uint64, error) {
	value, err := strconv.ParseUint(c.Query(paraName), common.BaseHex, common.BitSize64)
	if err != nil {
		return 0, fmt.Errorf("req int para [%s] is invalid", paraName)
	}
	return value, nil
}

func getStringReqPara(c *gin.Context, paraName string) (string, error) {
	value := c.Query(paraName)
	if value == "" {
		return "", fmt.Errorf("req string para [%s] is invalid", paraName)
	}
	return value, nil
}

type queryDispatcher struct {
	restfulmgr.GenericDispatcher
	name     string
	isString bool
}

func (query queryDispatcher) ParseData(c *gin.Context) (interface{}, error) {
	if query.isString {
		return getStringReqPara(c, query.name)
	} else {
		return getIntReqPara(c, query.name)
	}
}

type listDispatcher struct {
	restfulmgr.GenericDispatcher
}

func (list listDispatcher) ParseData(c *gin.Context) (interface{}, error) {
	return pageUtil(c)
}
