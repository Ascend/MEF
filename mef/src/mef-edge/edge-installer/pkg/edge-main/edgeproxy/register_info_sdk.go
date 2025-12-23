// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

package edgeproxy

import (
	"huawei.com/mindx/common/modulemgr"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/msgconv"
)

var staticResourceRouterListSdk = []modulemgr.RegisterModuleInfo{
	{Src: constants.ModEdgeOm, MsgOpt: constants.OptReport,
		MsgRes: constants.ResSoftwareVersion, ModuleName: constants.ModEdgeHub},
	{Src: constants.ModEdgeOm, MsgOpt: constants.OptGet,
		MsgRes: constants.ResDownloadCert, ModuleName: constants.ModEdgeHub},
	{Src: constants.ModEdgeOm, MsgOpt: constants.OptResponse,
		MsgRes: constants.ResPackLogResponse, ModuleName: constants.ModHandlerMgr},
}

func init() {
	forwardingRegisterInfoList = []msgconv.ForwardingRegisterInfo{
		{Source: msgconv.Edge, Operation: constants.OptPatch, Resource: constants.ResMefPodPatchPrefix,
			Event: msgconv.BeforeModification, Destination: constants.ModHandlerMgr},
	}
	staticResourceRouterList = append(staticResourceRouterList, staticResourceRouterListSdk...)
}
