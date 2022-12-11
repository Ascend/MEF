// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

func createApp(c *gin.Context) {
	res, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("create app: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.AppManagerName,
		Option:      common.Create,
		Resource:    common.App,
	}
	resp := common.SendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func queryApp(c *gin.Context) {
	appId, err := getReqAppId(c)
	if err != nil {
		hwlog.RunLog.Errorf("get app id failed: %s", err.Error())
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.AppManagerName,
		Option:      common.Query,
		Resource:    common.App,
	}

	resp := common.SendSyncMessageByRestful(appId, &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func updateApp(c *gin.Context) {
	res, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("update app: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.AppManagerName,
		Option:      common.Update,
		Resource:    common.App,
	}
	resp := common.SendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func listAppsInfo(c *gin.Context) {
	input, err := pageUtil(c)
	if err != nil {
		hwlog.OpLog.Error("list deployed apps: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, "", nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.AppManagerName,
		Option:      common.List,
		Resource:    common.App,
	}
	resp := common.SendSyncMessageByRestful(input, &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func deployApp(c *gin.Context) {
	res, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("deploy app: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, "", nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.AppManagerName,
		Option:      common.Deploy,
		Resource:    common.App,
	}
	resp := common.SendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func deleteApp(c *gin.Context) {
	res, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("delete app: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, "", nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.AppManagerName,
		Option:      common.Delete,
		Resource:    common.App,
	}
	resp := common.SendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func listAppInstance(c *gin.Context) {
	appId, err := getReqAppId(c)
	if err != nil {
		hwlog.OpLog.Error("list deployed app: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, "", nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.AppManagerName,
		Option:      common.List,
		Resource:    common.AppInstance,
	}
	resp := common.SendSyncMessageByRestful(appId, &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func listAppInstanceByNode(c *gin.Context) {
	res, err := getReqNodeId(c)
	if err != nil {
		hwlog.OpLog.Error("list deployed app: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, "", nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.AppManagerName,
		Option:      common.List,
		Resource:    common.AppInstanceByNode,
	}
	resp := common.SendSyncMessageByRestful(res, &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}
