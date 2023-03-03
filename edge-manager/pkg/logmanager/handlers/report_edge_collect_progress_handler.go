// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package handlers provides handlers to process business logic of log collection
package handlers

import (
	"encoding/json"

	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/logmanager/modules"
	"huawei.com/mindxedge/base/common/handlerbase"
	"huawei.com/mindxedge/base/common/logmgmt/logcollect"
	"huawei.com/mindxedge/base/modulemanager/model"
)

// GetReportEdgeProgressHandler get reportEdgeProgressHandler
func GetReportEdgeProgressHandler(taskMgr modules.TaskMgr) handlerbase.HandleBase {
	return &reportEdgeCollectProgressHandler{taskMgr: taskMgr}
}

type reportEdgeCollectProgressHandler struct {
	taskMgr modules.TaskMgr
}

func (h *reportEdgeCollectProgressHandler) Handle(msg *model.Message) error {
	hwlog.RunLog.Info("start to handle progress reporting")
	progress, err := h.parse(msg.Content)
	if err != nil {
		hwlog.RunLog.Errorf("failed to parse progress reporting message, %v", err)
		return err
	}
	if err := h.check(progress); err != nil {
		hwlog.RunLog.Errorf("failed to check progress reporting message, %v", err)
		return err
	}
	if err := h.taskMgr.NotifyProgress(progress, msg.GetNodeId()); err != nil {
		hwlog.RunLog.Errorf("failed to update progress, %v", err)
		return err
	}
	hwlog.RunLog.Info("handle progress reporting successful")
	return nil
}

func (h *reportEdgeCollectProgressHandler) parse(content interface{}) (logcollect.TaskProgress, error) {
	dataObj, ok := content.(map[string]interface{})
	if !ok {
		return logcollect.TaskProgress{}, nil
	}
	dataBytes, err := json.Marshal(dataObj)
	if err != nil {
		return logcollect.TaskProgress{}, nil
	}
	var progress logcollect.TaskProgress
	return progress, json.Unmarshal(dataBytes, &progress)
}

func (h *reportEdgeCollectProgressHandler) check(logcollect.TaskProgress) error {
	return nil
}

func (h *reportEdgeCollectProgressHandler) Parse(*model.Message) error {
	return nil
}

func (h *reportEdgeCollectProgressHandler) Check(*model.Message) error {
	return nil
}

func (h *reportEdgeCollectProgressHandler) PrintOpLogOk() {
}

func (h *reportEdgeCollectProgressHandler) PrintOpLogFail() {
}
