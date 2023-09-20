// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to  provide containerized application template management.
package appmanager

import (
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/appmanager/appchecker"
	"edge-manager/pkg/types"
	"edge-manager/pkg/util"
	"huawei.com/mindxedge/base/common"
)

// createTemplate create app template
func createTemplate(param interface{}) common.RespMsg {
	hwlog.RunLog.Info("start to create app template")
	var req CreateTemplateReq
	if err := common.ParamConvert(param, &req); err != nil {
		hwlog.RunLog.Error("create app template failed, error: request parameter convert failed")
		return common.RespMsg{Status: common.ErrorParamInvalid}
	}
	if checkResult := appchecker.NewCreateTemplateChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("app template create para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorCheckAppTemplateParams, Msg: checkResult.Reason, Data: nil}
	}
	if err := NewTemplateSupplementalChecker(req).Check(); err != nil {
		hwlog.RunLog.Errorf("app template create para check failed: %v", err)
		return common.RespMsg{Status: common.ErrorCheckAppTemplateParams,
			Msg: fmt.Sprintf("app template create para supplemental check failed: %v", err), Data: nil}
	}
	total, err := GetTableCount(AppTemplateDb{})
	if err != nil {
		hwlog.RunLog.Error("get app template num failed")
		return common.RespMsg{Status: common.ErrorCountAllTemplates, Msg: "get app template num failed", Data: nil}
	}

	if total >= MaxAppTemplate {
		errInfo := fmt.Sprintf("app template number has reached the upper limit[%v], can not create",
			MaxAppTemplate)
		hwlog.RunLog.Errorf(errInfo)
		return common.RespMsg{Status: common.ErrorExceedMaxTemplates, Msg: errInfo, Data: nil}
	}

	var template AppTemplateDb
	if err := req.ToDb(&template); err != nil {
		hwlog.RunLog.Errorf("create app template,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorCreateAppTemplate}
	}

	if err := RepositoryInstance().createTemplate(&template); err != nil {
		hwlog.RunLog.Errorf("create app template failed,err:%v", err)
		if strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
			return common.RespMsg{
				Status: common.ErrorCreateAppTemplate,
				Msg:    fmt.Sprintf("template with name %s already exists", template.TemplateName),
			}
		}
		return common.RespMsg{Status: common.ErrorCreateAppTemplate, Msg: "create app template in db failed"}
	}
	hwlog.RunLog.Info("create app template,success")
	return common.RespMsg{Status: common.Success, Data: template.ID}
}

// deleteTemplate delete app template
func deleteTemplate(param interface{}) common.RespMsg {
	hwlog.RunLog.Info("delete app template,start")
	req := DeleteTemplateReq{}
	if err := common.ParamConvert(param, &req); err != nil {
		hwlog.RunLog.Error("delete app template,failed,error:request parameter convert failed")
		return common.RespMsg{Status: common.ErrorParamConvert}
	}
	if checkResult := appchecker.NewDeleteTemplateChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("templates delete para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}
	var deleteRes types.BatchResp
	failedMap := make(map[string]string)
	deleteRes.FailedInfos = failedMap
	succeseedMap := make(map[uint64]bool)
	ids, err := RepositoryInstance().deleteTemplates(req.Ids)
	if err != nil {
		hwlog.RunLog.Errorf("delete app template failed,%v", err)
		return common.RespMsg{Status: common.ErrorDeleteAppTemplate, Msg: "delete app template failed"}
	}
	for _, id := range ids {
		deleteRes.SuccessIDs = append(deleteRes.SuccessIDs, id)
		succeseedMap[id] = true
	}

	for _, id := range req.Ids {
		if _, ok := succeseedMap[id]; !ok {
			errinfo := fmt.Sprintf("template db delete failed,id[%v] not exist", id)
			deleteRes.FailedInfos[strconv.Itoa(int(id))] = errinfo
		}
	}
	if len(deleteRes.FailedInfos) != 0 {
		return common.RespMsg{Status: common.ErrorDeleteAppTemplate, Msg: "", Data: deleteRes}
	}
	hwlog.RunLog.Info("delete app template,success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

// updateTemplate modify app template
func updateTemplate(param interface{}) common.RespMsg {
	hwlog.RunLog.Info("modify app template,start")
	var req UpdateTemplateReq
	if err := common.ParamConvert(param, &req); err != nil {
		hwlog.RunLog.Error("modify app template,failed,error:request parameter convert failed")
		return common.RespMsg{Status: common.ErrorParamInvalid}
	}
	if checkResult := appchecker.NewUpdateTemplateChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("app template update para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorCheckAppTemplateParams, Msg: checkResult.Reason, Data: nil}
	}
	if err := NewTemplateSupplementalChecker(req.CreateTemplateReq).Check(); err != nil {
		hwlog.RunLog.Errorf("app template create para check failed: %v", err)
		return common.RespMsg{Status: common.ErrorCheckAppTemplateParams,
			Msg: fmt.Sprintf("app template create para supplemental check failed: %v", err), Data: nil}
	}
	var template AppTemplateDb
	if err := req.ToDb(&template); err != nil {
		hwlog.RunLog.Errorf("create app template,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorModifyAppTemplate}
	}

	if err := RepositoryInstance().updateTemplate(&template); err != nil {
		hwlog.RunLog.Errorf("modify app template failed,error:%v", err.Error())
		return common.RespMsg{Status: common.ErrorModifyAppTemplate, Msg: err.Error()}
	}
	hwlog.RunLog.Info("modify app template,success")
	return common.RespMsg{Status: common.Success}
}

// getTemplates get app templates
func listTemplates(param interface{}) common.RespMsg {
	hwlog.RunLog.Info("list app templates, start")
	req, ok := param.(types.ListReq)
	if !ok {
		hwlog.RunLog.Error("list app templates,failed: para type is invalid")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "list app info error", Data: nil}
	}
	if checkResult := util.NewPaginationQueryChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("list app templates para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}

	templates, err := RepositoryInstance().getTemplates(req.Name, req.PageNum, req.PageSize)
	if err != nil {
		hwlog.RunLog.Errorf("list app templates,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorGetAppTemplates}
	}

	appTemplates := make([]AppTemplate, len(templates))
	for i, item := range templates {
		if err := (&appTemplates[i]).FromDb(&item); err != nil {
			hwlog.RunLog.Errorf("list app templates,failed,error:%v", err)
			return common.RespMsg{Status: common.ErrorGetAppTemplates}
		}
	}

	totalCount, err := RepositoryInstance().getTemplateCount(req.Name)
	if err != nil {
		hwlog.RunLog.Errorf("list app templates,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorGetAppTemplates}
	}
	var result ListTemplatesResp
	result.AppTemplates = appTemplates
	result.Total = totalCount
	hwlog.RunLog.Info("list app templates,success")
	return common.RespMsg{Status: common.Success, Data: result}
}

// getTemplate get app template detail
func getTemplate(param interface{}) common.RespMsg {
	hwlog.RunLog.Info("get app template detail,start")
	id, ok := param.(uint64)
	if !ok || id == 0 {
		hwlog.RunLog.Error("get app template failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get app template failed", Data: nil}
	}
	template, err := RepositoryInstance().getTemplate(id)
	if err == gorm.ErrRecordNotFound {
		hwlog.RunLog.Errorf("template id [%d] not found", id)
		return common.RespMsg{Status: common.ErrorTemplateNotFind,
			Msg: fmt.Sprintf("template id [%d] not found", id), Data: nil}
	}
	if err != nil {
		hwlog.RunLog.Errorf("get template id [%d] failed, db error", id)
		return common.RespMsg{Status: common.ErrorGetAppTemplates,
			Msg: fmt.Sprintf("get template id [%d] failed, db error", id), Data: nil}
	}
	var dto AppTemplate
	if err = (&dto).FromDb(template); err != nil {
		hwlog.RunLog.Errorf("get app template detail,failed,error:%v", err)
		return common.RespMsg{Status: common.ErrorGetAppTemplateDetail}
	}
	hwlog.RunLog.Info("get app template detail,success")
	return common.RespMsg{Status: common.Success, Data: dto}
}
