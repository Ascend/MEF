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
