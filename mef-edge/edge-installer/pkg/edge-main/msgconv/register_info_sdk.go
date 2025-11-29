// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.
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
