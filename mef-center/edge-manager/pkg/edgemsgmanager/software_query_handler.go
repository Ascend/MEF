// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package edgemsgmanager query the version of the edge downloaded software
package edgemsgmanager

import (
	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"
)

func queryEdgeSoftwareVersion(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start query edge software version")
	var serialNumber string
	if err := msg.ParseContent(&serialNumber); err != nil {
		hwlog.RunLog.Errorf("query edge software version failed: parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}

	if res := checker.GetRegChecker("",
		`^[a-zA-Z0-9]([-_a-zA-Z0-9]{0,62}[a-zA-Z0-9])?$`, true).Check(serialNumber); !res.Result {
		hwlog.RunLog.Errorf("check query software version para failed: %s", res.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: res.Reason, Data: nil}
	}

	nodeSoftwareInfo, err := getNodeSoftwareInfo(serialNumber)
	if err != nil {
		return common.RespMsg{Status: common.ErrorGetNodeSoftwareVersion, Msg: "", Data: nil}
	}

	return common.RespMsg{Status: common.Success, Msg: "", Data: nodeSoftwareInfo}
}
