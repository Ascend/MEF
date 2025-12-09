// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
