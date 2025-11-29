// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeproxy forward msg from device-om, edgecore or edge-om
package edgeproxy

import (
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
)

// MsgHandleFunc websocket server http handlers definition
type MsgHandleFunc func(msgBytes []byte)

// MsgDestination message destination type and name
type MsgDestination struct {
	DestType string // websocket conn or inner module
	DestName string
}

const (
	// MsgDestTypeModule message type sent to module
	MsgDestTypeModule = "inner_module"
	// MsgDestTypeWs message type sent to websocket
	MsgDestTypeWs = "websocket_conn"
)

// IsSyncMsgResp check if this msg is response to a sync request
func IsSyncMsgResp(msg *model.Message) bool {
	return msg.Header.IsSync && msg.Header.ParentId != ""
}

// IsInternalKubeedgeResp check if the message from edgecore/cloudcore is a response to edge-main's request
func IsInternalKubeedgeResp(msg *model.Message) bool {
	if msg.KubeEdgeRouter.Operation != constants.OptResponse {
		return false
	}
	return modulemgr.IsEnabledModule(msg.KubeEdgeRouter.Source) || msg.KubeEdgeRouter.Source == constants.ModEdgeMain
}
