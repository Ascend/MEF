// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

func createTemplate(c *gin.Context) {
	res, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("create template: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
	}
	resp := common.SendSyncMessageByRestful(string(res), &common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.AppManagerName,
		Option:      common.Create,
		Resource:    common.AppTemplate,
	})
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func updateTemplate(c *gin.Context) {
	res, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("modify template: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
	}
	resp := common.SendSyncMessageByRestful(string(res), &common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.AppManagerName,
		Option:      common.Update,
		Resource:    common.AppTemplate,
	})
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func deleteTemplate(c *gin.Context) {
	res, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("delete template: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
	}
	resp := common.SendSyncMessageByRestful(string(res), &common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.AppManagerName,
		Option:      common.Delete,
		Resource:    common.AppTemplate,
	})
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func getTemplates(c *gin.Context) {
	input, err := pageUtil(c)
	if err != nil {
		hwlog.OpLog.Error("list deployed apps: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, "", nil)
		return
	}

	resp := common.SendSyncMessageByRestful(input, &common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.AppManagerName,
		Option:      common.List,
		Resource:    common.AppTemplate,
	})
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func getTemplateDetail(c *gin.Context) {
	appId, err := getReqAppId(c)
	if err != nil {
		hwlog.RunLog.Errorf("get app id failed: %s", err.Error())
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
	}

	resp := common.SendSyncMessageByRestful(appId, &common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.AppManagerName,
		Option:      common.Get,
		Resource:    common.AppTemplate,
	})
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}
