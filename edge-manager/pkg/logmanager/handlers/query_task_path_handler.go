// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package handlers provides handlers to process business logic of log collection
package handlers

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/handlerbase"
	"huawei.com/mindxedge/base/common/logmgmt/logcollect"

	"edge-manager/pkg/logmanager/modules"
	"edge-manager/pkg/types"
)

// GetQueryTaskPathHandler get queryTaskPathHandler
func GetQueryTaskPathHandler(taskMgr modules.TaskMgr) handlerbase.HandleBase {
	return &queryTaskPathHandler{taskMgr: taskMgr}
}

type queryTaskPathHandler struct {
	taskMgr modules.TaskMgr
}

func (h *queryTaskPathHandler) Handle(msg *model.Message) error {
	hwlog.RunLog.Info("start to handle path query")
	req, err := h.parse(msg.Content)
	if err != nil {
		hwlog.RunLog.Errorf("failed to handle path query, %v", err)
		return sendResponse(common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error()}, msg)
	}
	if err := h.check(req); err != nil {
		hwlog.RunLog.Errorf("failed to handle path query, %v", err)
		return sendResponse(common.RespMsg{Status: common.ErrorParamInvalid, Msg: err.Error()}, msg)
	}
	var batchResp types.BatchResp
	failedMap := make(map[string]string)
	batchResp.FailedInfos = failedMap
	for _, node := range req.EdgeNodes {
		resp := logcollect.QueryTaskResp{
			Module:   req.Module,
			EdgeNode: node,
		}
		path, err := h.taskMgr.GetTaskPath(node)
		if err != nil {
			errInfo := fmt.Sprintf("failed to get path, %v", err)
			hwlog.RunLog.Error(errInfo)
			failedMap[node] = errInfo
			continue
		}
		resp.Data = path
		batchResp.SuccessIDs = append(batchResp.SuccessIDs, resp)
	}
	if len(batchResp.FailedInfos) > 0 {
		hwlog.RunLog.Error("failed to handle path query")
		return sendResponse(common.RespMsg{Status: common.ErrorLogCollectEdgeBusiness, Data: batchResp}, msg)
	}
	hwlog.RunLog.Info("handle path query successful")
	return sendResponse(common.RespMsg{Status: common.Success, Data: batchResp}, msg)
}

func (h *queryTaskPathHandler) parse(content interface{}) (logcollect.BatchQueryTaskReq, error) {
	var req logcollect.BatchQueryTaskReq
	return req, common.ParamConvert(content, &req)
}

func (h *queryTaskPathHandler) check(req logcollect.BatchQueryTaskReq) error {
	checkResult := getBatchQueryChecker().Check(req)
	if !checkResult.Result {
		return errors.New(checkResult.Reason)
	}
	return nil
}

func (h *queryTaskPathHandler) Parse(*model.Message) error {
	return nil
}

func (h *queryTaskPathHandler) Check(*model.Message) error {
	return nil
}

func (h *queryTaskPathHandler) PrintOpLogOk() {
}

func (h *queryTaskPathHandler) PrintOpLogFail() {
}
