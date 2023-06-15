// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager to manage node msg
package edgemsgmanager

import (
	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
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

	if res := checker.GetRegChecker("",
		`^[a-zA-Z0-9]([-_a-zA-Z0-9]{0,62}[a-zA-Z0-9])?$`, true).Check(serialNumber); !res.Result {
		hwlog.RunLog.Errorf("check download progress para failed: %s", res.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: res.Reason, Data: nil}
	}

	var processInfo types.ProgressInfo
	if nodeProgress, ok := nodesProgress[serialNumber]; ok {
		processInfo = nodeProgress
	}

	return common.RespMsg{Status: common.Success, Msg: "", Data: processInfo}
}
