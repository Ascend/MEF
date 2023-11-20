// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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
	msg.Content = map[string]interface{}{"taskId": ctx.Spec().Id, "module": "edgeNode"}
	response, err := modulemgr.SendSyncMessage(msg, sendMessageTimeout)
	if err != nil {
		hwlog.RunLog.Errorf("failed to send message to cloudhub, %v", err)
		utils.FeedbackTaskError(ctx, errors.New("failed to send message to cloudhub"))
		return
	}
	respMsg, ok := response.Content.(string)
	if !ok {
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
