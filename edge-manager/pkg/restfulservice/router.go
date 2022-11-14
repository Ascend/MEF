// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"edge-manager/module_manager"
	"edge-manager/module_manager/model"
	"edge-manager/pkg/common"
	"encoding/json"
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
}

func nodeRouter(engine *gin.Engine) {
	v1 := engine.Group("/edgemanager/v1/node")
	{
		v1.POST("/", createEdgeNode)
	}
}

func sendSyncMessageByRestful(input interface{}, router *router) (*model.Message, error) {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Error("create new message error")
		return nil, err
	}
	msg.SetRouter(router.source, router.destination, router.option, router.resource)
	msg.FillContent(input)
	respMsg, err := module_manager.SendSyncMessage(msg, common.ResponseTimeout)
	if err != nil {
		hwlog.RunLog.Error("get response error")
		return nil, err
	}
	return respMsg, nil
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
