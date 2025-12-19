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
	"regexp"
	"strings"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/handler"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-manager/pkg/constants"

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
	var taskId string
	if err := msg.ParseContent(&taskId); err != nil {
		hwlog.RunLog.Errorf("parse content failed: %v", err)
		return sendRestfulResponse(common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed"}, msg)
	}
	if matched, err := regexp.MatchString(constants.MultiNodesTaskIdRegexpStr, taskId); err != nil || !matched {
		hwlog.RunLog.Errorf("process task meet err: %v", err)
		return sendRestfulResponse(common.RespMsg{Status: common.ErrorParamConvert, Msg: "invalid task Id"}, msg)
	}

	taskCtx, err := taskschedule.DefaultScheduler().GetTaskContext(taskId)
	if err != nil {
		hwlog.RunLog.Errorf("failed to handle progress query, %v", err)
		return sendRestfulResponse(common.RespMsg{Status: common.ErrorLogDumpBusiness, Msg: err.Error()}, msg)
	}
	taskTree, err := taskCtx.GetSubTaskTree()
	if err != nil {
		hwlog.RunLog.Errorf("failed to handle progress query, %v", err)
		return sendRestfulResponse(common.RespMsg{Status: common.ErrorLogDumpBusiness, Msg: err.Error()}, msg)
	}
	hwlog.RunLog.Info("handle progress query successful")
	return sendRestfulResponse(common.RespMsg{Status: common.Success, Data: getTaskRespProgress(taskTree)}, msg)
}

func (h *queryProgressHandler) checkArgs(taskId string) (string, error) {
	if matched, err := regexp.MatchString(constants.MultiNodesTaskIdRegexpStr, taskId); err != nil || !matched {
		return "", errors.New("invalid task id")
	}
	return taskId, nil
}

func getTaskRespProgress(taskTree taskschedule.TaskTreeNode) QueryProgressResp {
	masterTaskStatus := taskTree.Current.Status
	reason := masterTaskStatus.Message
	if masterTaskStatus.Phase.IsFinished() && masterTaskStatus.Phase != taskschedule.Succeed {
		var subTaskReasons []string
		for _, subTask := range taskTree.Children {
			if !(subTask.Current.Status.Phase.IsFinished() && subTask.Current.Status.Phase != taskschedule.Succeed) {
				continue
			}
			var nodeID uint64
			if err := subTask.Current.Spec.Args.Get(constants.NodeID, &nodeID); err != nil {
				hwlog.RunLog.Error("failed to get node id")
			}
			subTaskReasons = append(subTaskReasons,
				fmt.Sprintf("(%d)%s", nodeID, subTask.Current.Status.Message))
		}
		if len(subTaskReasons) > 0 {
			reason = "failed to dump log for following nodes:" + strings.Join(subTaskReasons, ";")
		}
	}
	return QueryProgressResp{
		TaskId:     taskTree.Current.Spec.Id,
		Status:     masterTaskStatus.Phase,
		Reason:     reason,
		Progress:   masterTaskStatus.Progress,
		Data:       masterTaskStatus.Data,
		StartedAt:  NullableTime(masterTaskStatus.StartedAt),
		CreatedAt:  NullableTime(masterTaskStatus.CreatedAt),
		FinishedAt: NullableTime(masterTaskStatus.FinishedAt),
	}
}
