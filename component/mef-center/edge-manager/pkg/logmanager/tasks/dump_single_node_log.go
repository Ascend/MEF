// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tasks
package tasks

import (
	"errors"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/taskschedule"

	"edge-manager/pkg/constants"
	"edge-manager/pkg/logmanager/utils"
)

const (
	sendMessageTimeout = 5 * time.Second
)

// doDumpSingleNodeLog sends message to edge and tells edge to collect and upload logs
func doDumpSingleNodeLog(ctx taskschedule.TaskContext) {
	var serialNumber string
	if err := ctx.Spec().Args.Get(constants.NodeSnAndIp, &serialNumber); err != nil {
		hwlog.RunLog.Errorf("failed to parse serial number, %v", err)
		utils.FeedbackTaskError(ctx, errors.New("failed to parse serial number"))
		return
	}

	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("failed to create message, %v", err)
		utils.FeedbackTaskError(ctx, errors.New("failed to create message"))
		return
	}
	msg.SetRouter(constants.LogManagerName, common.CloudHubName, common.OptPost, constants.ResLogDumpTask)
	msg.SetNodeId(serialNumber)
	if err = msg.FillContent(map[string]interface{}{"taskId": ctx.Spec().Id, "module": "edgeNode"}); err != nil {
		hwlog.RunLog.Errorf("fill content failed: %v", err)
		utils.FeedbackTaskError(ctx, errors.New("failed to send message to cloudhub"))
		return
	}
	response, err := modulemgr.SendSyncMessage(msg, sendMessageTimeout)
	if err != nil {
		hwlog.RunLog.Errorf("failed to send message to cloudhub, %v", err)
		utils.FeedbackTaskError(ctx, errors.New("failed to send message to cloudhub"))
		return
	}
	var respMsg string
	if err = response.ParseContent(&respMsg); err != nil {
		hwlog.RunLog.Errorf("failed to parse response from cloudhub, %v", err)
		utils.FeedbackTaskError(ctx, errors.New("failed to parse response from cloudhub"))
		return
	}
	if respMsg != common.OK {
		hwlog.RunLog.Errorf("failed to send message to edge, the response cloudhub returns is %s", respMsg)
		utils.FeedbackTaskError(ctx, errors.New("failed to send message to edge"))
		return
	}

	hwlog.RunLog.Infof("send message to edge %s successful", serialNumber)
}
