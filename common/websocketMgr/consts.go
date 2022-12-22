// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package websocket this file for constants
package websocket

import "time"

// websocket default settings
const (
	handshakeTimeout         = 10 * time.Second
	readBufferSize           = 1024
	writeBufferSize          = 1024
	defaultReadLimit         = int64(1.5 * 1024 * 1024)
	defaultHeartbeatDuration = 5 * time.Second
	defaultHeartbeatTimeout  = 60 * time.Second

	// todo 待tls 公共能力实现，需要变成wss
	wsProtocol        = "ws://"
	serverPattern     = "/"
	defaultRetryCount = 5
	clientNameKey     = "clientName"
	retryTime         = 5 * time.Second
)
