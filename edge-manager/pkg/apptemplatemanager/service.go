// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package apptemplatemanager to  provide containerized application template management.
package apptemplatemanager

import (
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
	if err := req.Check(); err != nil {
		hwlog.RunLog.Errorf("create app template,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: err.Error()}
	}
	var template AppTemplateDb
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

// ModifyTemplate modify app template
func ModifyTemplate(param interface{}) common.RespMsg {
	hwlog.RunLog.Info("modify app template,start")
	var req AppTemplateDto
	if err := common.ParamConvert(param, &req); err != nil {
		hwlog.RunLog.Error("modify app template,failed,error:request parameter convert failed")
		return common.RespMsg{Status: common.ErrorParamInvalid}
	}
	if err := req.Check(); err != nil {
		hwlog.RunLog.Errorf("modify app template,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: err.Error()}
	}
	var template AppTemplateDb
	if err := req.ToDb(&template); err != nil {
		hwlog.RunLog.Errorf("create app template,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorModifyAppTemplate}
	}
	if err := RepositoryInstance().ModifyTemplate(&template); err != nil {
		hwlog.RunLog.Errorf("modify app template,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorModifyAppTemplate, Msg: err.Error()}
	}
	hwlog.RunLog.Info("modify app template,success")
	return common.RespMsg{Status: common.Success}
}

// GetTemplates get app templates
func GetTemplates(param interface{}) common.RespMsg {
	hwlog.RunLog.Info("get app templates,start")
	req := ReqGetTemplates{}
	if err := common.ParamConvert(param, &req); err != nil {
		hwlog.RunLog.Error("get app templates,failed,error:request parameter convert failed")
		return common.RespMsg{Status: common.ErrorParamInvalid}
	}
	templates, err := RepositoryInstance().GetTemplates(req.Name, req.PageNum, req.PageSize)
	if err != nil {
		hwlog.RunLog.Errorf("get app templates,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorGetAppTemplates, Msg: err.Error()}
	}
	result := make([]TemplateSummaryDto, len(templates))
	for i, item := range templates {
		(&result[i]).FromDb(&item)
	}
	hwlog.RunLog.Info("get app templates,success")
	return common.RespMsg{Status: common.Success, Data: result}
}

// GetTemplateDetail get app template detail
func GetTemplateDetail(param interface{}) common.RespMsg {
	hwlog.RunLog.Info("get app template detail,start")
	req := ReqGetTemplateDetail{}
	if err := common.ParamConvert(param, &req); err != nil {
		hwlog.RunLog.Error("get app template detail,failed,error:request parameter convert failed")
		return common.RespMsg{Status: common.ErrorParamInvalid}
	}
	template, err := RepositoryInstance().GetTemplate(req.Id)
	var dto AppTemplateDto
	if err = (&dto).FromDb(template); err != nil {
		hwlog.RunLog.Errorf("get app template detail,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorGetAppTemplateDetail}
	}
	hwlog.RunLog.Info("get app template detail,success")
	return common.RespMsg{Status: common.Success, Data: dto}
}
