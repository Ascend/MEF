// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager to manage node msg
package edgemsgmanager

import (
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager/model"
)

// UpdateEdgeSoftwareUpgradeProgress [method] update edge software upgrade progress
func UpdateEdgeSoftwareUpgradeProgress(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start to update node upgrade progress info")
	message, ok := input.(*model.Message)
	if !ok {
		hwlog.RunLog.Error("get message failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get message failed", Data: nil}
	}

	var req types.EdgeReportUpgradeResInfoReq
	if err := common.ParamConvert(message.GetContent(), &req); err != nil {
		hwlog.RunLog.Errorf("update node upgrade progress info error, %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "convert request error", Data: nil}
	}

	nodesProgress[req.SerialNumber] = req.ProgressInfo

	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}
