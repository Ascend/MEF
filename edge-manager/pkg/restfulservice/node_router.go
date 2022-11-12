// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"edge-manager/pkg/common"
	"edge-manager/pkg/util"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"
)

func createEdgeNode(c *gin.Context) {
	var reqContent util.CreateEdgeNodeReq
	if err := c.ShouldBindJSON(&reqContent); err != nil {
		hwlog.OpLog.Error("create node: convert request body failed")
		common.ConstructResp(c, common.ErrorParseBody, "", nil)
		return
	}
	router := router{
		source:      common.RestfulServiceName,
		destination: common.NodeManagerName,
		option:      common.Create,
		resource:    common.Node,
	}
	respMsg, err := sendSyncMessageByRestful(reqContent, &router)
	if err != nil {
		common.ConstructResp(c, common.ErrorsSendSyncMessageByRestful, "", nil)
		return
	}
	resp := marshalResponse(respMsg)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}
