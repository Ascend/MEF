// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"edge-manager/pkg/common"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"
)

func createEdgeNode(c *gin.Context) {
	res, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("create node: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
	}
	router := router{
		source:      common.RestfulServiceName,
		destination: common.NodeManagerName,
		option:      common.Create,
		resource:    common.Node,
	}
	resp := sendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func listNodeUnManaged(c *gin.Context) {
	input, err := pageUtil(c)
	if err != nil {
		hwlog.OpLog.Error("list node: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, "", nil)
		return
	}
	router := router{
		source:      common.RestfulServiceName,
		destination: common.NodeManagerName,
		option:      common.List,
		resource:    common.NodeUnManaged,
	}
	resp := sendSyncMessageByRestful(input, &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func listNodeManaged(c *gin.Context) {
	input, err := pageUtil(c)
	if err != nil {
		hwlog.OpLog.Error("list node: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, "", nil)
		return
	}
	router := router{
		source:      common.RestfulServiceName,
		destination: common.NodeManagerName,
		option:      common.List,
		resource:    common.Node,
	}
	resp := sendSyncMessageByRestful(input, &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func createEdgeNodeGroup(c *gin.Context) {
	reqContent, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("create node group: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
	}
	router := router{
		source:      common.RestfulServiceName,
		destination: common.NodeManagerName,
		option:      common.Create,
		resource:    common.NodeGroup,
	}
	resp := sendSyncMessageByRestful(string(reqContent), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func getNodeDetail(c *gin.Context) {
	res, err := common.BindUriWithJSON(c)
	if err != nil {
		hwlog.OpLog.Error("get node detail: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
	}
	router := router{
		source:      common.RestfulServiceName,
		destination: common.NodeManagerName,
		option:      common.Get,
		resource:    common.Node,
	}
	resp := sendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func modifyNode(c *gin.Context) {
	res, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("create node: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
	}
	router := router{
		source:      common.RestfulServiceName,
		destination: common.NodeManagerName,
		option:      common.Update,
		resource:    common.Node,
	}
	resp := sendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func getNodeStatistics(c *gin.Context) {
	router := router{
		source:      common.RestfulServiceName,
		destination: common.NodeManagerName,
		option:      common.Get,
		resource:    common.NodeStatistics,
	}
	resp := sendSyncMessageByRestful("", &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func listEdgeNodeGroup(c *gin.Context) {
	router := router{
		source:      common.RestfulServiceName,
		destination: common.NodeManagerName,
		option:      common.List,
		resource:    common.NodeGroup,
	}
	resp := sendSyncMessageByRestful("", &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func getEdgeNodeGroupDetail(c *gin.Context) {
	res, err := common.BindUriWithJSON(c)
	if err != nil {
		hwlog.OpLog.Error("get node group detail: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
	}
	router := router{
		source:      common.RestfulServiceName,
		destination: common.NodeManagerName,
		option:      common.Get,
		resource:    common.NodeGroup,
	}
	resp := sendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}
