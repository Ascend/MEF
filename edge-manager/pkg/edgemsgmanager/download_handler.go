// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager to manage node msg
package edgemsgmanager

import (
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

// downloadSoftware [method] down edge software
func downloadSoftware(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start update edge software")
	message, ok := input.(*model.Message)
	if !ok {
		hwlog.RunLog.Errorf("get message failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get message failed", Data: nil}
	}

	var req SoftwareDownloadInfo
	var err error
	if err = common.ParamConvert(message.GetContent(), &req); err != nil {
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}

	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
		return common.RespMsg{Status: common.ErrorNewMsg, Msg: "create message failed", Data: nil}
	}

	msg.SetRouter(common.NodeMsgManagerName, common.CloudHubName, common.OptPost, common.ResEdgeDownloadInfo)
	msg.FillContent(message.GetContent())
	var batchResp types.BatchResp
	for _, sn := range req.SerialNumbers {
		msg.SetNodeId(sn)

		err = modulemanager.SendMessage(msg)
		if err != nil {
			batchResp.FailedIDs = append(batchResp.FailedIDs, sn)
		} else {
			batchResp.SuccessIDs = append(batchResp.SuccessIDs, sn)
		}
	}

	if len(batchResp.FailedIDs) != 0 {
		hwlog.RunLog.Info("deal edge software upgrade info failed")
		return common.RespMsg{Status: common.ErrorSendMsgToNode, Msg: "", Data: batchResp}
	} else {
		hwlog.RunLog.Info("deal edge software upgrade info success")
		return common.RespMsg{Status: common.Success, Msg: "", Data: batchResp}
	}
}
