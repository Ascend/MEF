// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

//go:build MEFEdge_SDK

// Package handlers
package handlers

import (
	"huawei.com/mindx/common/modulemgr/handler"

	"edge-installer/pkg/common/constants"
)

func init() {
	registerInfoList = append(registerInfoList,
		handler.RegisterInfo{
			MsgOpt: constants.OptPost, MsgRes: constants.ResDumpLogTask, Handler: getDumpLogHandler()},
		handler.RegisterInfo{
			MsgOpt: constants.OptResponse, MsgRes: constants.ResPackLogResponse, Handler: new(packLogResultHandler)},
		handler.RegisterInfo{
			MsgOpt: constants.OptPatch, MsgRes: constants.ResMefPodPatchPrefix, Handler: new(podRestartEventHandler)},
	)
}
