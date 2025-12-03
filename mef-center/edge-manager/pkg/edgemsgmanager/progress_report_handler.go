// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package edgemsgmanager for edge report the software upgrade progress to center
package edgemsgmanager

import (
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/types"
)

// UpdateEdgeDownloadProgress [method] edge report the software upgrade progress to center
func UpdateEdgeDownloadProgress(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start to update node upgrade progress info")
	var req types.EdgeDownloadResInfo
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("update node upgrade progress info error: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}
	if checkResult := newDownloadResChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("check software download result failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}
	sn := msg.GetPeerInfo().Sn
	if err := nodesProgress.Set(sn, req.ProgressInfo, neverOverdue); err != nil {
		hwlog.RunLog.Errorf("set software download progress for %s failed: %v", sn, err)
		return common.RespMsg{Status: common.ErrorUpdateSoftwareDownloadProgress, Msg: "set cache error", Data: nil}
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}
