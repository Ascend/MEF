// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.
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
