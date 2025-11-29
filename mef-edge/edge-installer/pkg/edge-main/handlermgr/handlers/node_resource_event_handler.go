// Copyright (c)  2024. Huawei Technologies Co., Ltd.  All rights reserved.

// Package handlers
package handlers

import (
	"encoding/json"
	"errors"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
	"k8s.io/api/core/v1"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/msgconv/statusmanager"
)

type nodeResourceEventHandler struct {
}

// Handle handles node-resource related events
func (h nodeResourceEventHandler) Handle(msg *model.Message) error {
	err := h.handle(msg)
	response := constants.OK
	if err != nil {
		response = constants.Failed
	}
	if err := util.SendInnerMsgResponse(msg, response); err != nil {
		hwlog.RunLog.Errorf("failed to send sync response, %v", err)
	}
	return err
}

func (h nodeResourceEventHandler) handle(_ *model.Message) error {
	nodes, err := statusmanager.GetNodeStatusMgr().GetAll()
	if err != nil {
		return err
	}
	if len(nodes) != 1 {
		return errors.New("exactly one node expected")
	}
	var nodeStr string
	for _, str := range nodes {
		nodeStr = str
	}
	var node v1.Node
	if err := json.Unmarshal([]byte(nodeStr), &node); err != nil {
		return err
	}

	cpuValue, ok := node.Status.Capacity[v1.ResourceCPU]
	if !ok {
		return nil
	}
	if cpuValue.Sign() == 1 && !config.GetCapabilityCache().HasCapability(constants.CapabilityPodConfig) {
		config.GetCapabilityCache().Set(constants.CapabilityPodConfig, true)
		config.GetCapabilityCache().Set(constants.CapabilityAppTaskStop, true)
		config.GetCapabilityCache().Set(constants.CapabilityPodRestart, true)
		config.GetCapabilityCache().Set(constants.CapabilityPodResource, true)
		config.GetCapabilityCache().Set(constants.CapabilityUdpContainerPort, true)
		config.GetCapabilityCache().Set(constants.CapabilityResourceConfig, true)
		config.GetCapabilityCache().Notify()
	}
	return nil
}
