// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager effect edge software after upgrading
package edgemsgmanager

import (
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/types"
)

func upgradeEdgeSoftware(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start effect edge software")
	message, ok := input.(*model.Message)
	if !ok {
		hwlog.RunLog.Error("get message failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get message failed", Data: nil}
	}

	var req UpgradeSoftwareReq
	var err error
	if err = common.ParamConvert(message.GetContent(), &req); err != nil {
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}

	if checkResult := newUpgradeChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("check software upgrade para failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}
	upgradeConfig := message.GetContent()

	var batchResp types.BatchResp
	failedMap := make(map[string]string)
	batchResp.FailedInfos = failedMap
	for _, sn := range req.SerialNumbers {
		if err := sendUpgradeConfigToEdge(sn, upgradeConfig); err != nil {
			hwlog.RunLog.Error(err.Error())
			failedMap[sn] = err.Error()
			continue
		}

		batchResp.SuccessIDs = append(batchResp.SuccessIDs, sn)
	}

	if len(batchResp.FailedInfos) != 0 {
		hwlog.RunLog.Error("deal edge software upgrade info failed")
		return common.RespMsg{Status: common.ErrorSendMsgToNode, Msg: "", Data: batchResp}
	} else {
		hwlog.RunLog.Info("deal edge software upgrade info success")
		return common.RespMsg{Status: common.Success, Msg: "", Data: batchResp}
	}
}

func sendUpgradeConfigToEdge(sn string, upgradeConfig interface{}) error {
	msg, err := model.NewMessage()
	if err != nil {
		return fmt.Errorf("create message for %s failed", sn)
	}

	msg.SetRouter(common.NodeMsgManagerName, common.CloudHubName, common.OptPost, common.ResEdgeUpgradeInfo)
	msg.FillContent(upgradeConfig)
	msg.SetNodeId(sn)

	rsp, err := modulemgr.SendSyncMessage(msg, common.ResponseTimeout)
	if err != nil {
		return fmt.Errorf("send software upgrade info to %s failed", sn)
	}

	if content, ok := rsp.GetContent().(string); !ok || content != common.OK {
		return fmt.Errorf("mef edge process software upgrade info in %s failed", sn)
	}

	if err := nodesProgress.Set(sn, types.ProgressInfo{}, neverOverdue); err != nil {
		return fmt.Errorf("reset software download progress for %s failed: %v", sn, err)
	}

	return nil
}
