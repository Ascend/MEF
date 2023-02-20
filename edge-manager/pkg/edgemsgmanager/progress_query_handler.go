// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager to manage node msg
package edgemsgmanager

import (
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/types"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager/model"
)

func queryEdgeDownloadProgress(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start query edge software upgrade progress")
	message, ok := input.(*model.Message)
	if !ok {
		hwlog.RunLog.Error("get message failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get message failed", Data: nil}
	}

	serialNumber, ok := message.GetContent().(string)
	if !ok {
		hwlog.RunLog.Error("query edge software upgrade progress failed: para type not valid")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "query edge software upgrade progress" +
			" convert error", Data: nil}
	}

	var processInfo types.ProgressInfo
	if nodeProgress, ok := nodesProgress[serialNumber]; ok {
		processInfo = nodeProgress
	}

	return common.RespMsg{Status: common.Success, Msg: "", Data: processInfo}
}
