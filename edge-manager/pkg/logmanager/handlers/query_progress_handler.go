// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package handlers provides handlers to process business logic of log collection
package handlers

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/handler"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/taskschedule"
)

// NewQueryProgressHandler get queryProgressHandler
func NewQueryProgressHandler() handler.HandleBase {
	return &queryProgressHandler{}
}

type queryProgressHandler struct {
}

func (h *queryProgressHandler) Handle(msg *model.Message) error {
	hwlog.RunLog.Info("start to handle progress query")
	req, err := h.parseAndCheckArgs(msg.Content)
	if err != nil {
		hwlog.RunLog.Errorf("failed to handle progress query, %v", err)
		return sendRestfulResponse(common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error()}, msg)
	}
	taskCtx, err := taskschedule.DefaultScheduler().GetTaskContext(req)
	if err != nil {
		hwlog.RunLog.Errorf("failed to handle progress query, %v", err)
		return sendRestfulResponse(common.RespMsg{Status: common.ErrorLogDumpBusiness, Msg: err.Error()}, msg)
	}
	status, err := taskCtx.GetStatus()
	if err != nil {
		hwlog.RunLog.Errorf("failed to handle progress query, %v", err)
		return sendRestfulResponse(common.RespMsg{Status: common.ErrorLogDumpBusiness, Msg: err.Error()}, msg)
	}
	hwlog.RunLog.Info("handle progress query successful")
	return sendRestfulResponse(common.RespMsg{Status: common.Success, Data: getTaskRespProgress(req, status)}, msg)
}

func (h *queryProgressHandler) parseAndCheckArgs(content interface{}) (string, error) {
	req, ok := content.(string)
	if !ok {
		return "", errors.New("type error")
	}
	if checkResult := newQueryTaskProgressChecker().Check(req); !checkResult.Result {
		return "", fmt.Errorf("check request failed, %s", checkResult.Reason)
	}
	return req, nil
}

func getTaskRespProgress(taskId string, task taskschedule.TaskStatus) QueryProgressResp {
	return QueryProgressResp{
		TaskId:     taskId,
		Status:     task.Phase,
		Reason:     task.Message,
		Progress:   task.Progress,
		Data:       task.Data,
		StartedAt:  NullableTime(task.StartedAt),
		CreatedAt:  NullableTime(task.CreatedAt),
		FinishedAt: NullableTime(task.FinishedAt),
	}
}
