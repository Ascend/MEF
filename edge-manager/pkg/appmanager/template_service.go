// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to  provide containerized application template management.
package appmanager

import (
	"edge-manager/pkg/util"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

// CreateTemplate create app template
func CreateTemplate(param interface{}) common.RespMsg {
	hwlog.RunLog.Info("create app template,start")
	var req AppTemplateDto
	if err := common.ParamConvert(param, &req); err != nil {
		hwlog.RunLog.Error("create app template,failed,error:request parameter convert failed")
		return common.RespMsg{Status: common.ErrorParamInvalid}
	}

	var template AppTemplate
	if err := req.ToDb(&template); err != nil {
		hwlog.RunLog.Errorf("create app template,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorCreateAppTemplate}
	}
	if err := RepositoryInstance().CreateTemplate(&template); err != nil {
		hwlog.RunLog.Errorf("create app template,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorCreateAppTemplate, Msg: err.Error()}
	}
	hwlog.RunLog.Info("create app template,success")
	return common.RespMsg{Status: common.Success}
}

// DeleteTemplate delete app template
func DeleteTemplate(param interface{}) common.RespMsg {
	hwlog.RunLog.Info("delete app template,start")
	req := ReqDeleteTemplate{}
	if err := common.ParamConvert(param, &req); err != nil {
		hwlog.RunLog.Error("delete app template,failed,error:request parameter convert failed")
		return common.RespMsg{Status: common.ErrorParamInvalid}
	}
	if err := RepositoryInstance().DeleteTemplates(req.Ids); err != nil {
		hwlog.RunLog.Errorf("delete app template,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorDeleteAppTemplate, Msg: err.Error()}
	}
	hwlog.RunLog.Info("delete app template,success")
	return common.RespMsg{Status: common.Success}
}

// UpdateTemplate modify app template
func UpdateTemplate(param interface{}) common.RespMsg {
	hwlog.RunLog.Info("modify app template,start")
	var req AppTemplateDto
	if err := common.ParamConvert(param, &req); err != nil {
		hwlog.RunLog.Error("modify app template,failed,error:request parameter convert failed")
		return common.RespMsg{Status: common.ErrorParamInvalid}
	}

	var template AppTemplate
	if err := req.ToDb(&template); err != nil {
		hwlog.RunLog.Errorf("create app template,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorModifyAppTemplate}
	}
	if err := RepositoryInstance().UpdateTemplate(&template); err != nil {
		hwlog.RunLog.Errorf("modify app template,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorModifyAppTemplate, Msg: err.Error()}
	}
	hwlog.RunLog.Info("modify app template,success")
	return common.RespMsg{Status: common.Success}
}

// GetTemplates get app templates
func GetTemplates(param interface{}) common.RespMsg {
	hwlog.RunLog.Info("get app templates, start")
	req, ok := param.(util.ListReq)
	if !ok {
		hwlog.RunLog.Error("get app templates,failed: para type is invalid")
		return common.RespMsg{Status: "", Msg: "list app info error", Data: nil}
	}

	templates, err := RepositoryInstance().GetTemplates(req.Name, req.PageNum, req.PageSize)
	if err != nil {
		hwlog.RunLog.Errorf("get app templates,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorGetAppTemplates, Msg: err.Error()}
	}

	result := make([]AppTemplateDto, len(templates))
	for i, item := range templates {
		if err := (&result[i]).FromDb(&item); err != nil {
			hwlog.RunLog.Errorf("get app templates,failed,error:%v", err)
			return common.RespMsg{Status: common.ErrorGetAppTemplates, Msg: err.Error()}
		}
	}

	hwlog.RunLog.Info("get app templates,success")
	return common.RespMsg{Status: common.Success, Data: result}
}

// GetTemplateDetail get app template detail
func GetTemplateDetail(param interface{}) common.RespMsg {
	hwlog.RunLog.Info("get app template detail,start")
	id, ok := param.(uint64)
	if !ok {
		hwlog.RunLog.Error("get app template failed")
		return common.RespMsg{Status: "", Msg: "get app template failed", Data: nil}
	}
	template, err := RepositoryInstance().GetTemplate(id)
	var dto AppTemplateDto
	if err = (&dto).FromDb(template); err != nil {
		hwlog.RunLog.Errorf("get app template detail,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorGetAppTemplateDetail}
	}
	hwlog.RunLog.Info("get app template detail,success")
	return common.RespMsg{Status: common.Success, Data: dto}
}
