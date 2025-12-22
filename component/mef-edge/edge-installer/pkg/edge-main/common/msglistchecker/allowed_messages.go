// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package msglistchecker
package msglistchecker

import (
	"fmt"
	"regexp"

	"edge-installer/pkg/common/constants"
)

const (
	asyncMessage = false
	syncMessage  = true
	noParentID   = false
	hasParentID  = true
)

type messageRoute struct {
	source    string
	group     string
	operation string
}

type resourceInfo struct {
	hasParentID bool
	sync        bool
	resource    interface{}
}

var fdDownstreamAllowedRoutes = make(map[messageRoute][]resourceInfo)
var cloudCoreDownstreamAllowedRoutes = make(map[messageRoute][]resourceInfo)
var cloudCoreUpstreamAllowedRoutes = make(map[messageRoute][]resourceInfo)

func newMessageRoute(source, group, operation string) messageRoute {
	return messageRoute{source: source, group: group, operation: operation}
}

func newResourceInfo(hasParentID, isSync bool, resource interface{}) resourceInfo {
	return resourceInfo{hasParentID: hasParentID, sync: isSync, resource: resource}
}

func regexPattern(reg string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("^%s$", reg))
}

func initFdDownstreamMsgWhiteList() {
	controllerToResourceByUpdate := []resourceInfo{
		newResourceInfo(noParentID, syncMessage, regexPattern("websocket/pod/"+constants.FdPodNameRegex)),
		newResourceInfo(noParentID, syncMessage, regexPattern("websocket/configmap/"+constants.ConfigmapNameRegex)),
		newResourceInfo(noParentID, syncMessage, "websocket/secret/fusion-director-docker-registry-secret"),
	}
	controllerToResourceByDelete := []resourceInfo{
		newResourceInfo(noParentID, syncMessage, regexPattern("websocket/pod/"+constants.FdPodNameRegex)),
		newResourceInfo(noParentID, syncMessage, regexPattern("websocket/configmap/"+constants.ConfigmapNameRegex)),
		newResourceInfo(noParentID, asyncMessage, "websocket/pods_data"),
	}
	controllerToResourceByRestart := []resourceInfo{
		newResourceInfo(noParentID, syncMessage, regexPattern("websocket/pod/"+constants.FdPodNameRegex)),
	}
	controllerToHardwareByUpdate := []resourceInfo{
		newResourceInfo(noParentID, syncMessage, "websocket/container_info"),
		newResourceInfo(noParentID, syncMessage, "websocket/modelfiles"),
	}
	deviceOmToResourceByUpdate := []resourceInfo{
		newResourceInfo(noParentID, asyncMessage, "/edge/system/image-cert-info"),
	}
	deviceOmToHardwareByQuery := []resourceInfo{
		newResourceInfo(noParentID, asyncMessage, constants.QueryAllAlarm),
	}

	edgeManagerToHardwareByUpdate := []resourceInfo{
		newResourceInfo(noParentID, asyncMessage, "websocket/npu_sharing"),
	}

	fdDownstreamAllowedRoutes[newMessageRoute(constants.ControllerModule,
		constants.HardwareModule, constants.OptUpdate)] = controllerToHardwareByUpdate
	fdDownstreamAllowedRoutes[newMessageRoute(constants.ControllerModule,
		constants.ResourceModule, constants.OptUpdate)] = controllerToResourceByUpdate
	fdDownstreamAllowedRoutes[newMessageRoute(constants.ControllerModule,
		constants.ResourceModule, constants.OptDelete)] = controllerToResourceByDelete
	fdDownstreamAllowedRoutes[newMessageRoute(constants.ControllerModule,
		constants.ResourceModule, constants.OptRestart)] = controllerToResourceByRestart
	fdDownstreamAllowedRoutes[newMessageRoute(constants.DeviceOmModule,
		constants.ResourceModule, constants.OptUpdate)] = deviceOmToResourceByUpdate
	fdDownstreamAllowedRoutes[newMessageRoute(constants.DeviceOmModule,
		constants.HardwareModule, constants.OptQuery)] = deviceOmToHardwareByQuery
	fdDownstreamAllowedRoutes[newMessageRoute(constants.EdgeManagerModule,
		constants.HardwareModule, constants.OptUpdate)] = edgeManagerToHardwareByUpdate
}

func initCloudCoreDownstreamMsgWhiteList() {
	edgecontrollerToResourceByUpdate := []resourceInfo{
		newResourceInfo(noParentID, asyncMessage, regexPattern(constants.ResMefPodPrefix+constants.MefPodNameRegex)),
		newResourceInfo(noParentID, asyncMessage, constants.ResMefImagePullSecret),
	}

	edgecontrollerToResourceByDelete := []resourceInfo{
		newResourceInfo(noParentID, asyncMessage, regexPattern(constants.ResMefPodPrefix+constants.MefPodNameRegex)),
	}

	edgecontrollerToResourceByResponse := []resourceInfo{
		newResourceInfo(hasParentID, asyncMessage, constants.ResMefImagePullSecret),

		newResourceInfo(hasParentID, asyncMessage, regexPattern(constants.ResMefPodPrefix+constants.MefPodNameRegex)),
		newResourceInfo(hasParentID, asyncMessage, regexPattern(constants.ResMefPodPatchPrefix+constants.MefPodNameRegex)),
		newResourceInfo(hasParentID, asyncMessage, regexPattern(constants.ActionDefaultNodeStatus+constants.NodeNameRegx)),
		newResourceInfo(hasParentID, asyncMessage, regexPattern(constants.ActionDefaultNodePatch+constants.NodeNameRegx)),
		newResourceInfo(hasParentID, asyncMessage, regexPattern(constants.ResMefNodeLease+constants.NodeNameRegx)),
	}

	cloudCoreDownstreamAllowedRoutes[newMessageRoute(constants.EdgeControllerModule,
		constants.ResourceModule, constants.OptUpdate)] = edgecontrollerToResourceByUpdate

	cloudCoreDownstreamAllowedRoutes[newMessageRoute(constants.EdgeControllerModule,
		constants.ResourceModule, constants.OptDelete)] = edgecontrollerToResourceByDelete

	cloudCoreDownstreamAllowedRoutes[newMessageRoute(constants.EdgeControllerModule,
		constants.ResourceModule, constants.OptResponse)] = edgecontrollerToResourceByResponse
}

func initCloudCoreUpstreamMsgWhiteList() {
	edgedToMetaByUpdate := []resourceInfo{
		newResourceInfo(noParentID, asyncMessage, regexPattern(constants.ResMefNodeLease+constants.NodeNameRegx)),
	}
	edgedToMetaByQuery := []resourceInfo{
		newResourceInfo(noParentID, asyncMessage, regexPattern(constants.ActionDefaultNodeStatus+constants.NodeNameRegx)),
		newResourceInfo(noParentID, asyncMessage, regexPattern(constants.ResMefNodeLease+constants.NodeNameRegx)),
		newResourceInfo(noParentID, asyncMessage, constants.ResMefImagePullSecret),
	}
	edgedToMetaByInsert := []resourceInfo{
		newResourceInfo(noParentID, asyncMessage, regexPattern(constants.ActionDefaultNodeStatus+constants.NodeNameRegx)),
		newResourceInfo(noParentID, asyncMessage, regexPattern(constants.ResMefNodeLease+constants.NodeNameRegx)),
	}
	edgedToMetaByPatch := []resourceInfo{
		newResourceInfo(noParentID, asyncMessage, regexPattern(constants.ResMefPodPatchPrefix+constants.MefPodNameRegex)),
		newResourceInfo(noParentID, asyncMessage, regexPattern(constants.ActionDefaultNodePatch+constants.NodeNameRegx)),
	}
	edgedToMetaByDelete := []resourceInfo{
		newResourceInfo(noParentID, asyncMessage, regexPattern(constants.ResMefPodPrefix+constants.MefPodNameRegex)),
	}

	// response for response of query, edgecore -> cloudcore -> edgecore
	edgeControllerToResourceByResponse := []resourceInfo{
		newResourceInfo(hasParentID, asyncMessage, regexPattern(constants.ActionDefaultNodeStatus+constants.NodeNameRegx)),
		newResourceInfo(hasParentID, asyncMessage, regexPattern(constants.ResMefNodeLease+constants.NodeNameRegx)),
		newResourceInfo(hasParentID, asyncMessage, regexPattern(constants.ResMefPodPrefix+constants.MefPodNameRegex)),
		newResourceInfo(hasParentID, asyncMessage, constants.ResMefImagePullSecret),
	}

	nodeKeepalive := []resourceInfo{
		newResourceInfo(noParentID, asyncMessage, constants.ResourceTypeNode),
	}

	// edged -> meta
	cloudCoreUpstreamAllowedRoutes[newMessageRoute(constants.EdgedModule,
		constants.MetaModule, constants.OptUpdate)] = edgedToMetaByUpdate
	cloudCoreUpstreamAllowedRoutes[newMessageRoute(constants.EdgedModule,
		constants.MetaModule, constants.OptQuery)] = edgedToMetaByQuery
	cloudCoreUpstreamAllowedRoutes[newMessageRoute(constants.EdgedModule,
		constants.MetaModule, constants.OptInsert)] = edgedToMetaByInsert
	cloudCoreUpstreamAllowedRoutes[newMessageRoute(constants.EdgedModule,
		constants.MetaModule, constants.OptPatch)] = edgedToMetaByPatch
	cloudCoreUpstreamAllowedRoutes[newMessageRoute(constants.EdgedModule,
		constants.MetaModule, constants.OptDelete)] = edgedToMetaByDelete

	// edgecontroller -> resource
	cloudCoreUpstreamAllowedRoutes[newMessageRoute(constants.EdgeControllerModule,
		constants.ResourceModule, constants.OptResponse)] = edgeControllerToResourceByResponse

	// websocket -> resource
	cloudCoreUpstreamAllowedRoutes[newMessageRoute(constants.WebSocketModule,
		constants.ResourceModule, constants.OptKeepalive)] = nodeKeepalive

}

// because of message of module manager only use resource and option when routing,
// the key of mef-center <-> mef-edge message is only the option
var mefUpstreamAllowedRoutes = make(map[messageRoute][]resourceInfo)

func initMefUpstreamMsgWhiteList() {
	edgeToCenterByGet := []resourceInfo{
		newResourceInfo(noParentID, syncMessage, constants.ResConfig),
		newResourceInfo(noParentID, asyncMessage, constants.ResDownloadCert),
	}

	edgeToCenterByReport := []resourceInfo{
		newResourceInfo(noParentID, asyncMessage, constants.ResSoftwareVersion),
		newResourceInfo(noParentID, asyncMessage, constants.ResDownloadProgress),
		newResourceInfo(noParentID, asyncMessage, constants.ResDumpLogTaskError),
	}

	edgeToCenterByPost := []resourceInfo{
		newResourceInfo(noParentID, asyncMessage, constants.ResMefAlarmReport),
		newResourceInfo(noParentID, syncMessage, constants.ResEdgeCert),
	}

	edgeToCenterByResponse := []resourceInfo{
		newResourceInfo(noParentID, asyncMessage, constants.ResCertUpdate),
	}

	mefUpstreamAllowedRoutes[newMessageRoute("", "", constants.OptGet)] = edgeToCenterByGet
	mefUpstreamAllowedRoutes[newMessageRoute("", "", constants.OptReport)] = edgeToCenterByReport
	mefUpstreamAllowedRoutes[newMessageRoute("", "", constants.OptPost)] = edgeToCenterByPost
	mefUpstreamAllowedRoutes[newMessageRoute("", "", constants.OptResponse)] = edgeToCenterByResponse
}

func init() {
	// fd msg white list
	initFdDownstreamMsgWhiteList()
	// mef msg white list
	initCloudCoreDownstreamMsgWhiteList()
	initCloudCoreUpstreamMsgWhiteList()
	initMefUpstreamMsgWhiteList()
}
