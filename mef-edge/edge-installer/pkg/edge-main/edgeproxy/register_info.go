// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package edgeproxy define msg destination map
package edgeproxy

import (
	"fmt"
	"strings"

	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/msgconv"
)

var staticResourceMap = make(map[string]string)
var dynamicResourceMap = make(map[string]string)

var staticResourceRouterList = []modulemgr.RegisterModuleInfo{
	{Src: constants.ModDeviceOm, MsgOpt: constants.OptUpdate,
		MsgRes: constants.ResImageCertInfo, ModuleName: constants.ModEdgeOm},
	{Src: constants.ModDeviceOm, MsgOpt: constants.OptDelete,
		MsgRes: constants.ActionPodsData, ModuleName: constants.CfgRestore},
	{Src: constants.ModDeviceOm, MsgOpt: constants.OptUpdate,
		MsgRes: constants.ResNpuSharing, ModuleName: constants.ModEdgeOm},
	{Src: constants.ModDeviceOm, MsgOpt: constants.OptQuery,
		MsgRes: constants.QueryAllAlarm, ModuleName: constants.AlarmManager},
	{Src: constants.ModEdgeOm, MsgOpt: constants.OptUpdate,
		MsgRes: constants.ResConfigResult, ModuleName: constants.ModDeviceOm},
	{Src: constants.ModEdgeOm, MsgOpt: constants.OptUpdate, MsgRes: constants.ResAlarm,
		ModuleName: constants.AlarmManager},
	{Src: constants.ModDeviceOm, MsgOpt: constants.OptUpdate, MsgRes: constants.ActionSecret,
		ModuleName: constants.ModEdgeCore},
	{Src: constants.ModDeviceOm, MsgOpt: constants.OptUpdate, MsgRes: constants.ActionContainerInfo,
		ModuleName: constants.ModHandlerMgr},
	{Src: constants.ModDeviceOm, MsgOpt: constants.OptUpdate, MsgRes: constants.ActionModelFiles,
		ModuleName: constants.ModHandlerMgr},
	{Src: constants.ModEdgeCore, MsgOpt: constants.OptResponse, MsgRes: constants.ActionSecret,
		ModuleName: constants.ModDeviceOm},
	{Src: constants.ModEdgeOm, MsgOpt: constants.OptUpdate, MsgRes: constants.ResStatic,
		ModuleName: constants.ModDeviceOm},
}

var dynamicResourceRouterList = []modulemgr.RegisterModuleInfo{
	{Src: constants.ModDeviceOm, MsgOpt: constants.OptUpdate, MsgRes: constants.ActionConfigmap,
		ModuleName: constants.ModEdgeCore},
	{Src: constants.ModDeviceOm, MsgOpt: constants.OptDelete, MsgRes: constants.ActionConfigmap,
		ModuleName: constants.ModEdgeCore},
	{Src: constants.ModDeviceOm, MsgOpt: constants.OptUpdate, MsgRes: constants.ActionPod,
		ModuleName: constants.ModEdgeCore},
	{Src: constants.ModDeviceOm, MsgOpt: constants.OptDelete, MsgRes: constants.ActionPod,
		ModuleName: constants.ModEdgeCore},
	{Src: constants.ModDeviceOm, MsgOpt: constants.OptRestart, MsgRes: constants.ActionPod,
		ModuleName: constants.ModHandlerMgr},
	{Src: constants.ModEdgeCore, MsgOpt: constants.OptResponse, MsgRes: constants.ActionConfigmap,
		ModuleName: constants.ModDeviceOm},
	{Src: constants.ModEdgeCore, MsgOpt: constants.OptResponse, MsgRes: constants.ActionPod,
		ModuleName: constants.ModDeviceOm},
}

var forwardingRegisterInfoList = []msgconv.ForwardingRegisterInfo{
	{Source: msgconv.Edge, Operation: constants.OptPatch, Resource: constants.ActionPodPatch,
		Event: msgconv.BeforeModification, Destination: constants.ModHandlerMgr},
	{Source: msgconv.Edge, Operation: constants.OptInsert, Resource: constants.ActionDefaultNodeStatus,
		Event: msgconv.AfterDispatch, Destination: constants.ModHandlerMgr},
	{Source: msgconv.Edge, Operation: constants.OptPatch, Resource: constants.ActionDefaultNodePatch,
		Event: msgconv.AfterDispatch, Destination: constants.ModHandlerMgr},
}

// RegistryMsgRouters Registry msg routers
func RegistryMsgRouters() {
	for _, router := range staticResourceRouterList {
		staticResourceMap[router.Src+router.MsgOpt+router.MsgRes] = router.ModuleName
	}

	for _, router := range dynamicResourceRouterList {
		dynamicResourceMap[router.Src+router.MsgOpt+router.MsgRes] = router.ModuleName
	}
}

func getDestination(dest string) *MsgDestination {
	if dest == constants.ModEdgeCore || dest == constants.ModDeviceOm {
		return &MsgDestination{
			DestType: MsgDestTypeWs,
			DestName: dest,
		}
	}

	return &MsgDestination{
		DestType: MsgDestTypeModule,
		DestName: dest,
	}
}

// GetMsgDest GetByKey msg destination
func GetMsgDest(src string, msg *model.Message) (*MsgDestination, error) {
	if msg == nil {
		return nil, fmt.Errorf("invalid msg data")
	}

	resourceOperateKey := src + msg.GetOption() + msg.GetResource()
	if dest, ok := staticResourceMap[resourceOperateKey]; ok {
		return getDestination(dest), nil
	}

	for k, v := range dynamicResourceMap {
		if strings.HasPrefix(resourceOperateKey, k) {
			return getDestination(v), nil
		}
	}

	return nil, fmt.Errorf("msg destination not found")
}
