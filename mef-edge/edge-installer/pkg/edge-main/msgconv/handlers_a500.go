// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_A500

// Package msgconv
package msgconv

import (
	"encoding/json"
	"strings"
	"time"

	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"k8s.io/api/core/v1"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/job"
	"edge-installer/pkg/edge-main/msgconv/statusmanager"
)

func handlePodPatch(message *model.Message) error {
	metaKey := strings.ReplaceAll(message.KubeEdgeRouter.Resource, constants.ActionPodPatch, constants.ActionPod)
	data, err := statusmanager.GetPodStatusMgr().Get(metaKey)
	if err != nil {
		return err
	}
	var pod v1.Pod
	if err := json.Unmarshal([]byte(data), &pod); err != nil {
		return err
	}
	content := PodResp{Object: &pod}

	if err := respondToEdgeCore(message, content); err != nil {
		return err
	}
	return job.SyncPodStatus()
}

func handleNodeInsert(message *model.Message) error {
	var node v1.Node
	if err := message.ParseContent(&node); err != nil {
		return err
	}
	content := NodeResp{Object: &node}

	if err := respondToEdgeCore(message, content); err != nil {
		return err
	}
	return job.SyncNodeStatus()
}

func handleNodePatch(message *model.Message) error {
	metaKey := strings.ReplaceAll(
		message.Router.Resource, constants.ActionDefaultNodePatch, constants.ActionDefaultNodeStatus)
	data, err := statusmanager.GetNodeStatusMgr().Get(metaKey)
	if err != nil {
		return err
	}
	var node v1.Node
	if err := json.Unmarshal([]byte(data), &node); err != nil {
		return err
	}
	content := NodeResp{Object: &node}

	if err := respondToEdgeCore(message, content); err != nil {
		return err
	}
	return job.SyncNodeStatus()
}

func handleNodeQuery(message *model.Message) error {
	const defaultPodCIDR = "192.168.1.0/24"

	data, err := statusmanager.GetNodeStatusMgr().Get(message.KubeEdgeRouter.Resource)
	if err != nil {
		return err
	}
	var node v1.Node
	if err := json.Unmarshal([]byte(data), &node); err != nil {
		return err
	}
	// edgecores requires the `PodCIDR` and `PodCIDRs` fields
	node.Spec.PodCIDR = defaultPodCIDR
	node.Spec.PodCIDRs = []string{defaultPodCIDR}

	return respondToEdgeCore(message, node)
}

func respondToEdgeCore(request *model.Message, content interface{}) error {
	response, err := request.NewResponse()
	if err != nil {
		return err
	}

	response.Header.ParentID = request.Header.ID
	response.Header.ID = response.Header.Id
	response.Header.Timestamp = time.Now().UnixMilli()

	response.KubeEdgeRouter.Operation = constants.OptResponse
	response.KubeEdgeRouter.Group = constants.ResourceModule
	response.KubeEdgeRouter.Resource = request.KubeEdgeRouter.Resource
	response.KubeEdgeRouter.Source = constants.EdgeMain

	if err := response.FillContent(content); err != nil {
		return err
	}

	response.Router.Destination = constants.ModEdgeCore
	return modulemgr.SendAsyncMessage(response)
}
