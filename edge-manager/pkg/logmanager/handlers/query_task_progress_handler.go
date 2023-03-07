// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package handlers provides handlers to process business logic of log collection
package handlers

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/logmanager/modules"
	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/handlerbase"
	"huawei.com/mindxedge/base/common/logmgmt/logcollect"
	"huawei.com/mindxedge/base/modulemanager/model"
)

// GetQueryTaskProgressHandler get queryTaskProgressHandler
func GetQueryTaskProgressHandler(taskMgr modules.TaskMgr) handlerbase.HandleBase {
	return &queryTaskProgressHandler{taskMgr: taskMgr}
}

type queryTaskProgressHandler struct {
	taskMgr modules.TaskMgr
}

func (h *queryTaskProgressHandler) Handle(msg *model.Message) error {
	hwlog.RunLog.Info("start to handle progress query")
	req, err := h.parse(msg.Content)
	if err != nil {
		hwlog.RunLog.Errorf("failed to handle progress query, %v", err)
		return sendResponse(common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error()}, msg)
	}
	if err := h.check(req); err != nil {
		hwlog.RunLog.Errorf("failed to handle progress query, %v", err)
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
		progress, err := h.taskMgr.GetTaskProgress(node)
		if err != nil {
			errInfo := fmt.Sprintf("failed to get progress, %v", err)
			hwlog.RunLog.Error(errInfo)
			failedMap[node] = errInfo
			continue
		}
		resp.Data = progress
		batchResp.SuccessIDs = append(batchResp.SuccessIDs, resp)
	}
	if len(batchResp.FailedInfos) > 0 {
		hwlog.RunLog.Error("failed to handle progress query")
		return sendResponse(common.RespMsg{Status: common.ErrorLogCollectEdgeBusiness, Data: batchResp}, msg)
	}
	hwlog.RunLog.Info("handle progress query successful")
	return sendResponse(common.RespMsg{Status: common.Success, Data: batchResp}, msg)
}

func (h *queryTaskProgressHandler) parse(content interface{}) (logcollect.BatchQueryTaskReq, error) {
	var req logcollect.BatchQueryTaskReq
	return req, common.ParamConvert(content, &req)
}

func (h *queryTaskProgressHandler) check(req logcollect.BatchQueryTaskReq) error {
	checkResult := getBatchQueryChecker().Check(req)
	if !checkResult.Result {
		return errors.New(checkResult.Reason)
	}
	return nil
}

func (h *queryTaskProgressHandler) Parse(*model.Message) error {
	return nil
}

func (h *queryTaskProgressHandler) Check(*model.Message) error {
	return nil
}

func (h *queryTaskProgressHandler) PrintOpLogOk() {
}

func (h *queryTaskProgressHandler) PrintOpLogFail() {
}
