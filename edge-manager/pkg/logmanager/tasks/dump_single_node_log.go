// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package tasks
package tasks

import (
	"fmt"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-manager/pkg/constants"
	"edge-manager/pkg/logmanager/utils"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/taskschedule"
)

const (
	sendMessageTimeout = 5 * time.Second
)

// doDumpSingleNodeLog sends message to edge and tells edge to collect and upload logs
func doDumpSingleNodeLog(ctx taskschedule.TaskContext) {
	var serialNumber string
	if err := ctx.Spec().Args.Get(paramNameNodeSerialNumber, &serialNumber); err != nil {
		utils.FeedbackTaskError(ctx, fmt.Errorf("failed to parse serial number, %v", err))
		return
	}

	msg, err := model.NewMessage()
	if err != nil {
		utils.FeedbackTaskError(ctx, fmt.Errorf("failed to create message, %v", err))
		return
	}
	msg.SetRouter(constants.LogManagerName, common.CloudHubName, common.OptPost, constants.ResLogDumpTask)
	msg.SetNodeId(serialNumber)
	msg.Content = map[string]interface{}{"taskId": ctx.Spec().Id, "module": "edgeNode"}
	response, err := modulemgr.SendSyncMessage(msg, sendMessageTimeout)
	if err != nil {
		utils.FeedbackTaskError(ctx, fmt.Errorf("failed to send message to cloudhub, %v, node=%s", err, serialNumber))
		return
	}
	respMsg, ok := response.Content.(string)
	if !ok {
		utils.FeedbackTaskError(ctx, fmt.Errorf("failed to parse response from cloudhub, node=%s", serialNumber))
		return
	}
	if respMsg != common.OK {
		utils.FeedbackTaskError(ctx, fmt.Errorf("get unsuccessful response from cloudhub: %s", respMsg))
		return
	}

	hwlog.RunLog.Infof("send message to edge %s successful", serialNumber)
}
