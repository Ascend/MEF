// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_SDK

// Package handlermgr for deal every handler
package handlermgr

import (
	"huawei.com/mindx/common/modulemgr/handler"

	"edge-installer/pkg/common/constants"
)

func init() {
	registerInfoList = append(registerInfoList, []handler.RegisterInfo{
		{MsgOpt: constants.OptPost, MsgRes: constants.ResDownloadCert, Handler: new(saveCertHandlerSdk)},
		{MsgOpt: constants.OptGet, MsgRes: constants.InnerCert, Handler: new(getCertHandler)},
		{MsgOpt: constants.OptUpdate, MsgRes: constants.InnerPrepareDir, Handler: new(prepareDirHandler)},
		{MsgOpt: constants.OptPost, MsgRes: constants.ResPackLogRequest, Handler: new(packLogHandler)},
		{MsgOpt: constants.OptReport, MsgRes: constants.ResEdgeCloudConnection, Handler: new(cloudConnectHandler)},
	}...)
}
