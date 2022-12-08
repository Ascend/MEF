// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"encoding/json"
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
		Destination: common.TemplateManagerName,
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
		Destination: common.TemplateManagerName,
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
		Destination: common.TemplateManagerName,
		Option:      common.Delete,
		Resource:    common.AppTemplate,
	})
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func getTemplates(c *gin.Context) {
	handleUrlRequest(c, &common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.TemplateManagerName,
		Option:      common.List,
		Resource:    common.AppTemplate,
	})
}

func getTemplateDetail(c *gin.Context) {
	handleUrlRequest(c, &common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.TemplateManagerName,
		Option:      common.Get,
		Resource:    common.AppTemplate,
	})
}

func handleUrlRequest(c *gin.Context, route *common.Router) {
	param, err := json.Marshal(c.Request.URL.Query())
	if err != nil {
		common.ConstructResp(c, common.ErrorParseBody, "", nil)
		return
	}
	resp := common.SendSyncMessageByRestful(string(param), route)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}
