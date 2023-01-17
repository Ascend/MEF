// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"github.com/gin-gonic/gin"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/restfulmgr"
)

type appQueryDispatcher struct {
	restfulmgr.GenericDispatcher
}

func (app appQueryDispatcher) ParseData(c *gin.Context) common.Result {
	appId, err := getReqId(c, "appID")
	if err != nil {
		return common.Result{ResultFlag: false, ErrorMsg: "parse app id failed"}
	}

	return common.Result{ResultFlag: true, Data: appId}
}

type appListDispatcher struct {
	restfulmgr.GenericDispatcher
}

func (app appListDispatcher) ParseData(c *gin.Context) common.Result {
	input, err := pageUtil(c)
	if err != nil {
		return common.Result{ResultFlag: false, ErrorMsg: "parse app list para failed"}
	}

	return common.Result{ResultFlag: true, Data: input}
}

type appInstanceDispatcher struct {
	restfulmgr.GenericDispatcher
}

func (app appInstanceDispatcher) ParseData(c *gin.Context) common.Result {
	input, err := getReqNodeId(c)
	if err != nil {
		return common.Result{ResultFlag: false, ErrorMsg: "parse node id failed"}
	}

	return common.Result{ResultFlag: true, Data: input}
}

type listTemplateDispatcher struct {
	restfulmgr.GenericDispatcher
}

func (t listTemplateDispatcher) ParseData(c *gin.Context) common.Result {
	input, err := pageUtil(c)
	if err != nil {
		return common.Result{ResultFlag: false, ErrorMsg: "parse template list para failed"}
	}

	return common.Result{ResultFlag: true, Data: input}
}

type templateDetailDispatcher struct {
	restfulmgr.GenericDispatcher
}

func (t templateDetailDispatcher) ParseData(c *gin.Context) common.Result {
	templateId, err := getReqId(c, "id")
	if err != nil {
		return common.Result{ResultFlag: false, ErrorMsg: "parse template detail para failed"}
	}

	return common.Result{ResultFlag: true, Data: templateId}
}
