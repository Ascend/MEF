// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package edgemsgmanager effect edge software after upgrading
package edgemsgmanager

import (
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-manager/pkg/types"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/logmgmt"
)

func upgradeEdgeSoftware(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start effect edge software")
	var req UpgradeSoftwareReq
	var err error
	if err = msg.ParseContent(&req); err != nil {
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}

	if checkResult := newUpgradeChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("check software upgrade para failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}

	var batchResp types.BatchResp
	failedMap := make(map[string]string)
	batchResp.FailedInfos = failedMap
	for _, sn := range req.SerialNumbers {
		if err = sendUpgradeConfigToEdge(sn, msg.Content); err != nil {
			hwlog.RunLog.Errorf("send upgrade msg to edge failed: %v", err)
			failedMap[sn] = err.Error()
			continue
		}

		batchResp.SuccessIDs = append(batchResp.SuccessIDs, sn)
	}

	logmgmt.BatchOperationLog("send upgrade instruction to edge", batchResp.SuccessIDs)
	if len(batchResp.FailedInfos) != 0 {
		hwlog.RunLog.Error("deal edge software upgrade info failed")
		return common.RespMsg{Status: common.ErrorSendMsgToNode, Msg: "", Data: batchResp}
	} else {
		hwlog.RunLog.Info("deal edge software upgrade info success")
		return common.RespMsg{Status: common.Success, Msg: "", Data: batchResp}
	}
}

func sendUpgradeConfigToEdge(sn string, upgradeConfig []byte) error {
	msg, err := model.NewMessage()
	if err != nil {
		return fmt.Errorf("create message for %s failed", sn)
	}

	msg.SetRouter(common.NodeMsgManagerName, common.CloudHubName, common.OptPost, common.ResEdgeUpgradeInfo)
	if err = msg.FillContent(upgradeConfig); err != nil {
		return fmt.Errorf("fill content failed: %v", err)
	}
	msg.SetNodeId(sn)

	rsp, err := modulemgr.SendSyncMessage(msg, common.ResponseTimeout)
	if err != nil {
		return fmt.Errorf("send msg failed: %v", err)
	}

	var content string
	if err = rsp.ParseContent(&content); err != nil {
		return fmt.Errorf("parse resp failed: %v", err)
	}
	if content != common.OK {
		return fmt.Errorf("mef edge process software upgrade info in %s failed", sn)
	}

	if err := nodesProgress.Set(sn, types.ProgressInfo{}, neverOverdue); err != nil {
		return fmt.Errorf("reset software download progress for %s failed: %v", sn, err)
	}

	return nil
}
