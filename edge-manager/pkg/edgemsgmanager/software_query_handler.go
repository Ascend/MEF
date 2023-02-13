// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager to manage node msg
package edgemsgmanager

import (
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager/model"
)

// queryEdgeSoftwareVersion [method] query edge software version
func queryEdgeSoftwareVersion(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start query edge software version")
	message, ok := input.(*model.Message)
	if !ok {
		hwlog.RunLog.Error("get message failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get message failed", Data: nil}
	}

	serialNumber, ok := message.GetContent().(string)
	if !ok {
		hwlog.RunLog.Error("query edge software version failed: para type not valid")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "query edge software version " +
			"request convert error", Data: nil}
	}

	nodeSoftwareInfo, err := getNodeSoftwareInfo(serialNumber)
	if err != nil {
		return common.RespMsg{Status: common.ErrorGetNodeSoftwareVersion, Msg: "", Data: nil}
	}

	return common.RespMsg{Status: common.Success, Msg: "", Data: nodeSoftwareInfo}
}
