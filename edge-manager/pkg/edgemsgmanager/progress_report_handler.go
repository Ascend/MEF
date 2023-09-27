// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager for edge report the software upgrade progress to center
package edgemsgmanager

import (
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/types"
)

// UpdateEdgeDownloadProgress [method] edge report the software upgrade progress to center
func UpdateEdgeDownloadProgress(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start to update node upgrade progress info")
	message, ok := input.(*model.Message)
	if !ok {
		hwlog.RunLog.Error("get message failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get message failed", Data: nil}
	}

	var req types.EdgeReportUpgradeResInfoReq
	if err := common.ParamConvert(message.GetContent(), &req); err != nil {
		hwlog.RunLog.Errorf("update node upgrade progress info error: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "convert request error", Data: nil}
	}

	if err := nodesProgress.Set(req.SerialNumber, req.ProgressInfo, neverOverdue); err != nil {
		hwlog.RunLog.Errorf("set software download progress for %s failed: %v", req.SerialNumber, err)
		return common.RespMsg{Status: common.ErrorUpdateSoftwareDownloadProgress, Msg: "set cache error", Data: nil}
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}
