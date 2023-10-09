// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package cloudhub server init
package cloudhub

import (
	"huawei.com/mindx/common/websocketmgr"

	"edge-manager/pkg/constants"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
)

const (
	alarmHandlerRate     = 35
	alarmHandlerCapacity = 90

	defaultHandlerRate     = 16
	defaultHandlerCapacity = 2048
)

var regInfoList = []websocketmgr.RegisterModuleInfo{
	{MsgOpt: common.OptGet, MsgRes: common.ResConfig, ModuleName: common.NodeMsgManagerName,
		MsgRate: defaultHandlerRate, MsgCapacity: defaultHandlerCapacity},
	{MsgOpt: common.OptReport, MsgRes: common.ResDownloadProgress, ModuleName: common.NodeMsgManagerName,
		MsgRate: defaultHandlerRate, MsgCapacity: defaultHandlerCapacity},
	{MsgOpt: common.OptReport, MsgRes: common.ResSoftwareInfo, ModuleName: common.NodeManagerName,
		MsgRate: defaultHandlerRate, MsgCapacity: defaultHandlerCapacity},
	{MsgOpt: common.OptGet, MsgRes: common.ResDownLoadCert, ModuleName: common.NodeMsgManagerName,
		MsgRate: defaultHandlerRate, MsgCapacity: defaultHandlerCapacity},
	{MsgOpt: common.OptPost, MsgRes: common.ResEdgeCert, ModuleName: common.CloudHubName,
		MsgRate: defaultHandlerRate, MsgCapacity: defaultHandlerCapacity},
	{MsgOpt: common.OptResp, MsgRes: common.CertWillExpired, ModuleName: common.CertUpdaterName,
		MsgRate: defaultHandlerRate, MsgCapacity: defaultHandlerCapacity},
	{MsgOpt: common.OptPost, MsgRes: requests.ReportAlarmRouter, ModuleName: common.AlarmManagerName,
		MsgRate: alarmHandlerRate, MsgCapacity: alarmHandlerCapacity},
	{MsgOpt: common.OptReport, MsgRes: constants.ResLogDumpError, ModuleName: constants.LogManagerName,
		MsgRate: defaultHandlerRate, MsgCapacity: defaultHandlerCapacity},
}

func getRegModuleInfoList() []websocketmgr.RegisterModuleInfo {
	return regInfoList
}
