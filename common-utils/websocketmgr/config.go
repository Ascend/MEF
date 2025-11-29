// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

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
