// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"edge-manager/module_manager"
	"edge-manager/module_manager/model"
	"edge-manager/pkg/common"
	"edge-manager/pkg/util"
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"
)

type router struct {
	source      string
	destination string
	option      string
	resource    string
}

func setRouter(engine *gin.Engine) {
	engine.Use(gin.Recovery())
	nodeRouter(engine)
	appRouter(engine)
	wsRouter(engine)
}

func nodeRouter(engine *gin.Engine) {
	node := engine.Group("/edgemanager/v1/node")
	{
		node.POST("/", createEdgeNode)
		node.GET("/list/managed", listNodeManaged)
		node.GET("/list/unmanaged", listNodeUnManaged)
	}
	nodeGroup := engine.Group("/edgemanager/v1/nodegroup")
	{
		nodeGroup.POST("/", createEdgeNodeGroup)
	}
}

func appRouter(engine *gin.Engine) {
	app := engine.Group("/edgemanager/v1/app")
	{
		app.POST("/", createApp)
		app.GET("/list", listApp)

	}
}

func wsRouter(engine *gin.Engine) {
	v1 := engine.Group("/edgemanager/v1/ws")
	{
		v1.POST("/", upgradeSfw)
	}
}

func sendSyncMessageByRestful(input interface{}, router *router) common.RespMsg {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Error("new message error")
		return common.RespMsg{Status: common.ErrorsSendSyncMessageByRestful, Msg: "", Data: nil}
	}
	msg.SetRouter(router.source, router.destination, router.option, router.resource)
	msg.FillContent(input)
	respMsg, err := module_manager.SendSyncMessage(msg, common.ResponseTimeout)
	if err != nil {
		hwlog.RunLog.Error("get response error")
		return common.RespMsg{Status: common.ErrorsSendSyncMessageByRestful, Msg: "", Data: nil}
	}
	return marshalResponse(respMsg)
}

func marshalResponse(respMsg *model.Message) common.RespMsg {
	content := respMsg.GetContent()
	respStr, err := json.Marshal(content)
	if err != nil {
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "", Data: nil}
	}
	var resp common.RespMsg
	if err := json.Unmarshal(respStr, &resp); err != nil {
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "", Data: nil}
	}
	return resp
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
