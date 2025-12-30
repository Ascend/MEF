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

// Package msgconv
package msgconv

import (
	"k8s.io/api/core/v1"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/msgconv/securitysetters"
)

var msgconvHandlers = []messageHandler{
	{operation: constants.OptPatch, resource: constants.ActionDefaultNodePatch, source: Edge,
		contentType: "",
		setters:     []setter{setNodePatchRequest}},
	{operation: constants.OptResponse, resource: constants.ActionDefaultNodePatch, source: Cloud,
		contentType: NodeResp{},
		setters:     []setter{setNodePatchResponse}},

	{operation: constants.OptInsert, resource: constants.ActionDefaultNodeStatus, source: Edge,
		contentType: v1.Node{},
		setters:     []setter{setNodeInsertRequest}},
	{operation: constants.OptResponse, resource: constants.ActionDefaultNodeStatus, source: Cloud,
		contentType: map[string]interface{}{},
		setters:     []setter{setNodeResponse}},

	{operation: constants.OptPatch, resource: constants.ResMefPodPatchPrefix, source: Edge,
		contentType: "",
		setters:     []setter{setPodPatchRequest}},
	{operation: constants.OptResponse, resource: constants.ResMefPodPatchPrefix, source: Cloud,
		contentType: PodResp{},
		setters:     []setter{setPodPatchResponse}},

	{operation: constants.OptUpdate, resource: constants.ResMefPodPrefix, source: Cloud, contentType: v1.Pod{},
		setters: []setter{securitysetters.SetPodUpdate, setPodUpdateRequest}},
	// trigger deletion for pod records
	{operation: constants.OptDelete, resource: constants.ResMefPodPrefix, source: Cloud, contentType: v1.Pod{}},
}
