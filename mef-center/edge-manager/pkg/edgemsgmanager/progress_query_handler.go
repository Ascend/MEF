// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager query edge software download progress
package edgemsgmanager

import (
	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/types"
)

func queryEdgeDownloadProgress(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start query edge software upgrade progress")
	var serialNumber string
	if err := msg.ParseContent(&serialNumber); err != nil {
		hwlog.RunLog.Errorf("query edge software upgrade progress failed: parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}

	if res := checker.GetRegChecker("",
		`^[a-zA-Z0-9]([-_a-zA-Z0-9]{0,62}[a-zA-Z0-9])?$`, true).Check(serialNumber); !res.Result {
		hwlog.RunLog.Errorf("check download progress para failed: %s", res.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: res.Reason, Data: nil}
	}

	value, err := nodesProgress.Get(serialNumber)
	if err != nil {
		hwlog.RunLog.Errorf("get download progress for %s failed: %v", serialNumber, err)
		return common.RespMsg{Status: common.ErrorGetSoftwareDownloadProgress,
			Msg: "get download progress failed", Data: nil}
	}

	processInfo, ok := value.(types.ProgressInfo)
	if !ok {
		hwlog.RunLog.Errorf("convert download progress for %s failed", serialNumber)
		return common.RespMsg{Status: common.ErrorGetSoftwareDownloadProgress, Msg: "type convert error", Data: nil}
	}

	return common.RespMsg{Status: common.Success, Msg: "", Data: processInfo}
}
