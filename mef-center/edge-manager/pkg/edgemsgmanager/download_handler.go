// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package edgemsgmanager deal edge software download info and send to edge
package edgemsgmanager

import (
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/utils"

	"edge-manager/pkg/types"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/logmgmt"
)

// downloadSoftware [method] download software to edge
func downloadSoftware(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start deal edge software download info")
	var req SoftwareDownloadInfo
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}
	if req.DownloadInfo.Password != nil {
		defer utils.ClearSliceByteMemory(*(req.DownloadInfo.Password))
	}
	if checkResult := NewDownloadChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("check software download para failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}
	resp, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Error("create message failed")
		return common.RespMsg{Status: common.ErrorNewMsg, Msg: "create message failed", Data: nil}
	}
	resp.SetRouter(common.NodeMsgManagerName, common.CloudHubName, common.OptPost, common.ResEdgeDownloadInfo)
	if err = resp.FillContent(msg.Content); err != nil {
		hwlog.RunLog.Errorf("fill content failed: %v", err)
		return common.RespMsg{Status: common.ErrorNewMsg, Msg: "fill content failed", Data: nil}
	}
	batchResp := sendDownloadInfo(req, resp)
	if len(batchResp.FailedInfos) != 0 {
		hwlog.RunLog.Error("deal edge software upgrade info failed")
		return common.RespMsg{Status: common.ErrorSendMsgToNode, Msg: "", Data: batchResp}
	} else {
		hwlog.RunLog.Info("deal edge software download info success")
		return common.RespMsg{Status: common.Success, Msg: "", Data: batchResp}
	}
}

func sendDownloadInfo(req SoftwareDownloadInfo, msg *model.Message) types.BatchResp {
	var batchResp types.BatchResp
	failedMap := make(map[string]string)
	batchResp.FailedInfos = failedMap
	for _, sn := range req.SerialNumbers {
		if err := nodesProgress.Set(sn, types.ProgressInfo{}, neverOverdue); err != nil {
			hwlog.RunLog.Errorf("set software download progress for %s failed: %v", sn, err)
			failedMap[sn] = fmt.Sprintf("set software download progress failed: %v", err)
			continue
		}
		msg.SetNodeId(sn)
		hwlog.RunLog.Errorf("start to send msg to %v", msg.Router.Destination)
		rsp, err := modulemgr.SendSyncMessage(msg, common.ResponseTimeout)
		if err != nil {
			hwlog.RunLog.Errorf("send software download info to %s failed: %v", sn, err)
			failedMap[sn] = fmt.Sprintf("send software download info failed: %v", err)
			continue
		}
		var content string
		if err = rsp.ParseContent(&content); err != nil {
			hwlog.RunLog.Errorf("parse mef edge process software download info in %s failed: %v", sn, err)
			failedMap[sn] = "parse mef edge process software download info failed"
			continue
		}
		if content != common.OK {
			hwlog.RunLog.Errorf("parse mef edge process software download info in %s failed", sn)
			failedMap[sn] = "parse mef edge process software download info failed"
			continue
		}
		batchResp.SuccessIDs = append(batchResp.SuccessIDs, sn)
	}
	logmgmt.BatchOperationLog("send software download msg to edge", batchResp.SuccessIDs)
	return batchResp
}
