// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to  provide containerized application template management.
package appmanager

import (
	"edge-manager/pkg/util"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

// createTemplate create app template
func createTemplate(param interface{}) common.RespMsg {
	hwlog.RunLog.Info("create app template,start")
	var req CreateTemplateReq
	if err := common.ParamConvert(param, &req); err != nil {
		hwlog.RunLog.Error("create app template,failed,error:request parameter convert failed")
		return common.RespMsg{Status: common.ErrorParamInvalid}
	}

	checker := templateParaChecker{req: &req}
	if err := checker.Check(); err != nil {
		hwlog.RunLog.Errorf("template create para check failed: %s", err.Error())
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	var template AppTemplateDb
	if err := req.ToDb(&template); err != nil {
		hwlog.RunLog.Errorf("create app template,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorCreateAppTemplate}
	}

	if err := RepositoryInstance().createTemplate(&template); err != nil {
		hwlog.RunLog.Errorf("create app template,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorCreateAppTemplate, Msg: err.Error()}
	}
	hwlog.RunLog.Info("create app template,success")
	return common.RespMsg{Status: common.Success}
}

// deleteTemplate delete app template
func deleteTemplate(param interface{}) common.RespMsg {
	hwlog.RunLog.Info("delete app template,start")
	req := DeleteTemplateReq{}
	if err := common.ParamConvert(param, &req); err != nil {
		hwlog.RunLog.Error("delete app template,failed,error:request parameter convert failed")
		return common.RespMsg{Status: common.ErrorParamInvalid}
	}
	if err := RepositoryInstance().deleteTemplates(req.Ids); err != nil {
		hwlog.RunLog.Errorf("delete app template,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorDeleteAppTemplate, Msg: err.Error()}
	}
	hwlog.RunLog.Info("delete app template,success")
	return common.RespMsg{Status: common.Success}
}

// updateTemplate modify app template
func updateTemplate(param interface{}) common.RespMsg {
	hwlog.RunLog.Info("modify app template,start")
	var req UpdateTemplateReq
	if err := common.ParamConvert(param, &req); err != nil {
		hwlog.RunLog.Error("modify app template,failed,error:request parameter convert failed")
		return common.RespMsg{Status: common.ErrorParamInvalid}
	}

	checker := templateParaChecker{req: &req.CreateTemplateReq}
	if err := checker.Check(); err != nil {
		hwlog.RunLog.Errorf("template update para check failed: %s", err.Error())
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	var template AppTemplateDb
	if err := req.ToDb(&template); err != nil {
		hwlog.RunLog.Errorf("create app template,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorModifyAppTemplate}
	}

	if err := RepositoryInstance().updateTemplate(&template); err != nil {
		hwlog.RunLog.Errorf("modify app template,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorModifyAppTemplate, Msg: err.Error()}
	}
	hwlog.RunLog.Info("modify app template,success")
	return common.RespMsg{Status: common.Success}
}

// getTemplates get app templates
func getTemplates(param interface{}) common.RespMsg {
	hwlog.RunLog.Info("get app templates, start")
	req, ok := param.(util.ListReq)
	if !ok {
		hwlog.RunLog.Error("get app templates,failed: para type is invalid")
		return common.RespMsg{Status: "", Msg: "list app info error", Data: nil}
	}

	templates, err := RepositoryInstance().getTemplates(req.Name, req.PageNum, req.PageSize)
	if err != nil {
		hwlog.RunLog.Errorf("get app templates,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorGetAppTemplates, Msg: err.Error()}
	}

	appTemplates := make([]AppTemplate, len(templates))
	for i, item := range templates {
		if err := (&appTemplates[i]).FromDb(&item); err != nil {
			hwlog.RunLog.Errorf("get app templates,failed,error:%v", err)
			return common.RespMsg{Status: common.ErrorGetAppTemplates, Msg: err.Error()}
		}
	}

	totalCount, err := RepositoryInstance().getTemplateCount(req.Name)
	if err != nil {
		hwlog.RunLog.Errorf("get app templates,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorGetAppTemplates, Msg: err.Error()}
	}

	var result ListTemplatesResp
	result.AppTemplates = appTemplates
	result.Total = totalCount
	hwlog.RunLog.Info("get app templates,success")
	return common.RespMsg{Status: common.Success, Data: result}
}

// getTemplate get app template detail
func getTemplate(param interface{}) common.RespMsg {
	hwlog.RunLog.Info("get app template detail,start")
	id, ok := param.(uint64)
	if !ok {
		hwlog.RunLog.Error("get app template failed")
		return common.RespMsg{Status: "", Msg: "get app template failed", Data: nil}
	}
	template, err := RepositoryInstance().getTemplate(id)
	var dto AppTemplate
	if err = (&dto).FromDb(template); err != nil {
		hwlog.RunLog.Errorf("get app template detail,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorGetAppTemplateDetail}
	}
	hwlog.RunLog.Info("get app template detail,success")
	return common.RespMsg{Status: common.Success, Data: dto}
}
