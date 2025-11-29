// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_SDK

package innerclient

import (
	"huawei.com/mindx/common/modulemgr"

	"edge-installer/pkg/common/constants"
)

var registerInfoListSdk = []*modulemgr.RegisterModuleInfo{
	{MsgOpt: constants.OptReport, MsgRes: constants.ResEdgeCloudConnection, ModuleName: constants.ModEdgeOm},
	{MsgOpt: constants.OptReport, MsgRes: constants.InnerSoftwareVersion, ModuleName: constants.UpgradeManagerName},
	{MsgOpt: constants.OptPost, MsgRes: constants.InnerSoftwareVerification, ModuleName: constants.UpgradeManagerName},
	{MsgOpt: constants.OptPost, MsgRes: constants.ResUpgradeInfo, ModuleName: constants.UpgradeManagerName},
	{MsgOpt: constants.OptPost, MsgRes: constants.ResPackLogRequest, ModuleName: constants.ModEdgeOm},
	{MsgOpt: constants.OptPost, MsgRes: constants.ResDownloadCert, ModuleName: constants.ModEdgeOm},
	{MsgOpt: constants.OptUpdate, MsgRes: constants.InnerPrepareDir, ModuleName: constants.ModEdgeOm},
	{MsgOpt: constants.OptGet, MsgRes: constants.InnerCert, ModuleName: constants.ModEdgeOm},
}

func init() {
	registerInfoList = append(registerInfoList, registerInfoListSdk...)
}
