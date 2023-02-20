// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"errors"
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
			Destination: common.AppManagerName}, "appID"},
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
			Destination:  common.AppManagerName}, "appID"},
		queryDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/node",
			Method:       http.MethodGet,
			Destination:  common.AppManagerName}, "nodeID"},
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
			Destination:  common.AppManagerName}, "configmapID"},
		listDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/configmap/list",
			Method:       http.MethodGet,
			Destination:  common.AppManagerName}},
	},
}

var configRouterDispatchers = map[string][]restfulmgr.DispatcherItf{
	"/edgemanager/v1/image": {
		restfulmgr.GenericDispatcher{
			RelativePath: "/config",
			Method:       http.MethodPost,
			Destination:  common.ConfigManagerName},
		restfulmgr.GenericDispatcher{
			RelativePath: "/update",
			Method:       http.MethodPost,
			Destination:  common.ConfigManagerName},
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
			Destination: common.AppManagerName}, "id"},
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
	restfulmgr.InitRouter(engine, configRouterDispatchers)
	restfulmgr.InitRouter(engine, edgeAccountRouterDispatchers)
	restfulmgr.InitRouter(engine, softwareRouterDispatchers)
	connCertRouter(engine)
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
			Destination: common.NodeManagerName}, "id"},
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
		capDispatcher{restfulmgr.GenericDispatcher{
			RelativePath: "/capability",
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
			Destination: common.NodeManagerName}, "id"},
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

var edgeAccountRouterDispatchers = map[string][]restfulmgr.DispatcherItf{
	"/edgemanager/v1/edge-account": {
		restfulmgr.GenericDispatcher{
			Method:      http.MethodPost,
			Destination: common.EdgeInstallerName},
	},
}

var softwareRouterDispatchers = map[string][]restfulmgr.DispatcherItf{
	"/edgemanager/v1/software": {
		restfulmgr.GenericDispatcher{
			RelativePath: "/update",
			Method:       http.MethodPost,
			Destination:  common.EdgeInstallerName},
	},
}

func connCertRouter(engine *gin.Engine) {
	v1 := engine.Group("/edgemanager/v1/cert")
	{
		v1.POST("/download", downloadCert)
	}
}

func pageUtil(c *gin.Context) (types.ListReq, error) {
	input := types.ListReq{}
	var err error
	// for slice page on ucd
	input.PageNum, err = getIntReq(c, "pageNum")
	if err != nil {
		return input, err
	}
	input.PageSize, err = getIntReq(c, "pageSize")
	if err != nil {
		return input, err
	}
	// for fuzzy query
	input.Name = c.Query("name")
	return input, nil
}

func getIntReq(c *gin.Context, idName string) (uint64, error) {
	value, err := strconv.ParseUint(c.Query(idName), common.BaseHex, common.BitSize64)
	if err != nil {
		return 0, fmt.Errorf("id name [%s] is invalid", idName)
	}
	return value, nil
}

func getCapReq(c *gin.Context) (types.CapReq, error) {
	input := types.CapReq{}
	// nodeID is an optional parameter
	if c.Query("nodeID") == "" {
		hwlog.RunLog.Info("the nodeID parameter of get capability is empty")
		return input, nil
	}
	var err error
	input.NodeID, err = getIntReq(c, "nodeID")
	if err != nil {
		hwlog.RunLog.Errorf("get nodeID failed, error:%s", err)
		return input, err
	}
	if input.NodeID <= 0 {
		hwlog.RunLog.Error("nodeID is an invalid value")
		return input, errors.New("nodeID is an invalid value")
	}
	return input, nil
}

type queryDispatcher struct {
	restfulmgr.GenericDispatcher
	name string
}

func (query queryDispatcher) ParseData(c *gin.Context) (interface{}, error) {
	return getIntReq(c, query.name)
}

type listDispatcher struct {
	restfulmgr.GenericDispatcher
}

func (list listDispatcher) ParseData(c *gin.Context) (interface{}, error) {
	return pageUtil(c)
}

type capDispatcher struct {
	restfulmgr.GenericDispatcher
}

func (node capDispatcher) ParseData(c *gin.Context) (interface{}, error) {
	return getCapReq(c)
}
