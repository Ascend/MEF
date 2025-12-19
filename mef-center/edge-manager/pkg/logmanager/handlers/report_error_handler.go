// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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

	"edge-manager/pkg/constants"
)

// NewReportErrorHandler get reportErrorHandler
func NewReportErrorHandler() handler.HandleBase {
	return &reportErrorHandler{}
}

type reportErrorHandler struct {
	ip           string
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
	hwlog.RunLog.Errorf("receive error message from edge(%s): %s", h.serialNumber, h.errInfo.Message)
	status := taskschedule.TaskStatus{Phase: taskschedule.Failed, Reason: h.errInfo.Reason, Message: h.errInfo.Message}
	if err := taskCtx.UpdateStatus(status); err != nil {
		hwlog.RunLog.Errorf("failed to update progress, %v", err)
		return err
	}
	hwlog.RunLog.Info("handle progress report successful")
	return nil
}

func (h *reportErrorHandler) Parse(msg *model.Message) error {
	h.serialNumber = msg.GetPeerInfo().Sn
	ip, err := h.getIpBySn(h.serialNumber)
	if err != nil {
		hwlog.RunLog.Errorf("get ip by sn failed, error: %v", err)
		return errors.New("get ip by sn failed")
	}
	h.ip = ip
	h.errInfo = TaskErrorInfo{}
	return msg.ParseContent(&h.errInfo)
}

func (h *reportErrorHandler) Check(*model.Message) error {
	if checkResult := newTaskErrorChecker().Check(h.errInfo); !checkResult.Result {
		return fmt.Errorf("check error info failed, %v", checkResult.Reason)
	}
	return nil
}

func (h *reportErrorHandler) PrintOpLogOk() {
	hwlog.OpLog.Infof("[edge(%s)@%s] %s %s success", h.serialNumber, h.ip, common.OptReport, constants.ResLogDumpError)
}

func (h *reportErrorHandler) PrintOpLogFail() {
	hwlog.OpLog.Errorf("[edge(%s)@%s] %s %s failed", h.serialNumber, h.ip, common.OptReport, constants.ResLogDumpError)
}

func (h *reportErrorHandler) getIpBySn(sn string) (string, error) {
	router := common.Router{
		Source:      constants.LogManagerName,
		Destination: common.NodeManagerName,
		Option:      common.Get,
		Resource:    common.GetIpBySn,
	}
	resp := common.SendSyncMessageByRestful(sn, &router, common.ResponseTimeout)
	if resp.Status != common.Success {
		hwlog.RunLog.Errorf("get ip by sn through node manager failed, error: %s", resp.Msg)
		return "", errors.New("get ip by sn through node manager failed")
	}
	ip, ok := resp.Data.(string)
	if !ok {
		hwlog.RunLog.Error("resp data type from node manager is invalid")
		return "", errors.New("resp data type from node manager is invalid")
	}
	return ip, nil
}
