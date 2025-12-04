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
