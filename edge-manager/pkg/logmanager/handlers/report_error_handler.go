// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package handlers provides handlers to process business logic of log collection
package handlers

import (
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/handler"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/taskschedule"

	"edge-manager/pkg/constants"
)

// NewReportErrorHandler get reportErrorHandler
func NewReportErrorHandler() handler.HandleBase {
	return &reportErrorHandler{}
}

type reportErrorHandler struct {
	serialNumber string
	errInfo      TaskErrorInfo
}

func (h *reportErrorHandler) Handle(msg *model.Message) error {
	hwlog.RunLog.Info("start to handle progress report")
	taskCtx, err := taskschedule.DefaultScheduler().GetTaskContext(h.errInfo.Id)
	if err != nil {
		hwlog.RunLog.Errorf("failed to update progress, %v", err)
		return err
	}
	status := taskschedule.TaskStatus{Phase: taskschedule.Failed, Reason: h.errInfo.Reason, Message: h.errInfo.Message}
	if err := taskCtx.UpdateStatus(status); err != nil {
		hwlog.RunLog.Errorf("failed to update progress, %v", err)
		return err
	}
	hwlog.RunLog.Info("handle progress report successful")
	return nil
}

func (h *reportErrorHandler) Parse(msg *model.Message) error {
	h.serialNumber = msg.GetNodeId()
	return utils.ObjectConvert(msg.Content, &h.errInfo)
}
func (h *reportErrorHandler) Check(*model.Message) error {
	if checkResult := newTaskErrorChecker().Check(h.errInfo); !checkResult.Result {
		return fmt.Errorf("check error info failed, %v", checkResult.Reason)
	}
	return nil
}

func (h *reportErrorHandler) PrintOpLogOk() {
	hwlog.OpLog.Errorf("edge(%s) %s %s failed", h.serialNumber, common.OptReport, constants.ResLogDumpError)
}

func (h *reportErrorHandler) PrintOpLogFail() {
	hwlog.OpLog.Infof("edge(%s) %s %s success", h.serialNumber, common.OptReport, constants.ResLogDumpError)
}
