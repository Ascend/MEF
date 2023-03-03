// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager to manage node msg
package edgemsgmanager

import (
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

// downloadSoftware [method] down edge software
func downloadSoftware(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start deal edge software download info")
	message, ok := input.(*model.Message)
	if !ok {
		hwlog.RunLog.Error("get message failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get message failed", Data: nil}
	}

	var req SoftwareDownloadInfo
	var err error
	if err = common.ParamConvert(message.GetContent(), &req); err != nil {
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}

	if checkResult := newDownloadChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("check software download para failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}

	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Error("create message failed")
		return common.RespMsg{Status: common.ErrorNewMsg, Msg: "create message failed", Data: nil}
	}

	msg.SetRouter(common.NodeMsgManagerName, common.CloudHubName, common.OptPost, common.ResEdgeDownloadInfo)
	msg.FillContent(message.GetContent())
	var batchResp types.BatchResp
	failedMap := make(map[string]string)
	batchResp.FailedInfos = failedMap
	for _, sn := range req.SerialNumbers {
		msg.SetNodeId(sn)

		rsp, err := modulemanager.SendSyncMessage(msg, common.ResponseTimeout)
		if err != nil {
			errInfo := fmt.Sprintf("send software download info to %s failed", sn)
			hwlog.RunLog.Error(errInfo)
			failedMap[sn] = errInfo
			continue
		}

		if content, ok := rsp.GetContent().(string); !ok || content != common.OK {
			errInfo := fmt.Sprintf("mef edge process software download info in %s failed", sn)
			hwlog.RunLog.Error(errInfo)
			failedMap[sn] = errInfo
			continue
		}

		batchResp.SuccessIDs = append(batchResp.SuccessIDs, sn)
		nodesProgress[sn] = types.ProgressInfo{}
	}

	if len(batchResp.FailedInfos) != 0 {
		hwlog.RunLog.Error("deal edge software upgrade info failed")
		return common.RespMsg{Status: common.ErrorSendMsgToNode, Msg: "", Data: batchResp}
	} else {
		hwlog.RunLog.Info("deal edge software download info success")
		return common.RespMsg{Status: common.Success, Msg: "", Data: batchResp}
	}
}
