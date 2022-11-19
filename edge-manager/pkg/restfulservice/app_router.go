// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"edge-manager/pkg/common"
	"github.com/gin-gonic/gin"

	"huawei.com/mindx/common/hwlog"
)

func createApp(c *gin.Context) {
	res, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("create app: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
	}
	router := router{
		source:      common.RestfulServiceName,
		destination: common.AppManagerName,
		option:      common.Create,
		resource:    common.App,
	}
	resp := sendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func listAppsDeployed(c *gin.Context) {
	input, err := pageUtil(c)
	if err != nil {
		hwlog.OpLog.Error("list deployed apps: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, "", nil)
		return
	}
	router := router{
		source:      common.RestfulServiceName,
		destination: common.AppManagerName,
		option:      common.List,
		resource:    common.App,
	}
	resp := sendSyncMessageByRestful(input, &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}
