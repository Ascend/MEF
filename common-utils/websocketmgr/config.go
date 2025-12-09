// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package websocketmgr

import (
	"time"

	"github.com/gorilla/websocket"
)

// websocket default settings
const (
	defaultHandshakeTimeout  = 10 * time.Second
	wsReadBufferSize         = 1024
	wsWriteBufferSize        = 1024
	defaultReadSizeLimit     = int64(1.5 * 1024 * 1024)
	defaultHeartbeatInterval = 5 * time.Second
	defaultHeartbeatTimeout  = 60 * time.Second

	wssProtocol             = "wss://"
	svcUrl                  = "/"
	clientNameKey           = "clientName"
	realIpKey               = "X-Real-IP"
	retryInterval           = 5 * time.Second
	reconnectInterval       = 3 * time.Second
	maxTryConnInterval      = 128 * time.Second
	tryConnIntervalGrowRate = 2
	defaultReadTimeout      = 30 * time.Second
	defaultWriteTimeout     = 30 * time.Second
	defaultHeaderSizeLimit  = 1024
	defaultConnLimit        = 1024
	WsMsgTypeText           = websocket.TextMessage
	WsMsgTypeBinary         = websocket.BinaryMessage
	WsMsgTypeClose          = websocket.CloseMessage
	WsMsgTypePing           = websocket.PingMessage
	WsMsgTypePong           = websocket.PongMessage
)

// isValidWsMsgType make sure WebSocket message type is valid
func isValidWsMsgType(msgType int) bool {
	return msgType == WsMsgTypeText ||
		msgType == WsMsgTypeBinary ||
		msgType == WsMsgTypePing ||
		msgType == WsMsgTypePong ||
		msgType == WsMsgTypeClose
}
