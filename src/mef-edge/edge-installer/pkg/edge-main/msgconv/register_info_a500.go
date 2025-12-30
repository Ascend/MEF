// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_A500

// Package msgconv
package msgconv

import (
	"k8s.io/api/core/v1"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/types"
	"edge-installer/pkg/edge-main/msgconv/securitysetters"
)

var msgconvHandlers = []messageHandler{
	{operation: constants.OptUpdate, resource: constants.ActionSecret, source: Cloud, contentType: v1.Secret{},
		setters: []setter{securitysetters.SetSecretUpdate, setKind, setSourceToEdgeController}},
	{operation: constants.OptResponse, resource: constants.ActionSecret, source: Edge, contentType: "",
		setters: []setter{setSourceToController}},

	{operation: constants.OptUpdate, resource: constants.ActionConfigmap, source: Cloud, contentType: v1.ConfigMap{},
		setters: []setter{securitysetters.SetConfigmapUpdate, setKind, setSourceToEdgeController}},
	{operation: constants.OptDelete, resource: constants.ActionConfigmap, source: Cloud, contentType: v1.ConfigMap{},
		setters: []setter{securitysetters.SetConfigmapDelete, setKind, setSourceToEdgeController}},
	{operation: constants.OptResponse, resource: constants.ActionConfigmap, source: Edge, contentType: "",
		setters: []setter{setSourceToController}},

	{operation: constants.OptUpdate, resource: constants.ActionPod, source: Cloud, contentType: v1.Pod{},
		setters: []setter{securitysetters.SetPodUpdate,
			setKind, setSourceToEdgeController, setAsync, setPodSpecForUpdate}},
	{operation: constants.OptDelete, resource: constants.ActionPod, source: Cloud, contentType: v1.Pod{},
		setters: []setter{securitysetters.SetPodDelete, setKind, setSourceToEdgeController, setPodSpecForDelete}},
	{operation: constants.OptRestart, resource: constants.ActionPod, source: Cloud, contentType: "",
		setters: []setter{setPodRestartRoute}},
	{operation: constants.OptResponse, resource: constants.ActionPod, source: Edge, contentType: "",
		setters: []setter{setSourceToController}},
	{operation: constants.OptPatch, resource: constants.ActionPodPatch, source: Edge,
		handleFunc: handlePodPatch},

	{operation: constants.OptInsert, resource: constants.ActionDefaultNodeStatus, source: Edge, contentType: v1.Node{},
		setters: []setter{setKind}, handleFunc: handleNodeInsert},
	{operation: constants.OptQuery, resource: constants.ActionDefaultNodeStatus, source: Edge,
		handleFunc: handleNodeQuery},
	{operation: constants.OptPatch, resource: constants.ActionDefaultNodePatch, source: Edge,
		handleFunc: handleNodePatch},

	{operation: constants.OptUpdate, resource: constants.ActionModelFileInfo, source: Cloud,
		contentType: types.ModelFileInfo{},
		setters:     []setter{securitysetters.SetModelFileUpdate}},
}
