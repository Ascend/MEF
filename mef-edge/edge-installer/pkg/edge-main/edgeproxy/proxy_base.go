// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeproxy forward msg from device-om, edgecore or edge-om
package edgeproxy

import (
	"fmt"
	"io"
	"time"

	"github.com/gorilla/websocket"
	"huawei.com/mindx/common/hwlog"
)

type proxyInterface interface {
	Start(conn *websocket.Conn) error
}

type proxyBase struct {
	lastAlive time.Time
}

func (pb *proxyBase) readMsg(conn *websocket.Conn) ([]byte, error) {
	messageType, reader, err := conn.NextReader()
	if err != nil {
		return nil, fmt.Errorf("websocket NextReader for error: %v", err)
	}
	pb.lastAlive = time.Now()
	if messageType != websocket.TextMessage {
		hwlog.RunLog.Errorf("received unsupported message type: [%v]", messageType)
		return nil, nil
	}
	msgBytes, err := io.ReadAll(io.LimitReader(reader, defaultReadLimitInBytes))
	if err != nil {
		return nil, fmt.Errorf("read websocket message error: %v", err)
	}
	return msgBytes, nil
}
