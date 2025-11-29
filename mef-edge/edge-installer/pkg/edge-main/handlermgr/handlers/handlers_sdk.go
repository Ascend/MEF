// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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
