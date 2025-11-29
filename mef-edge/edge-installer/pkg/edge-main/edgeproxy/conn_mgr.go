// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeproxy forward msg from device-om, edgecore or edge-om
package edgeproxy

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common"
)

// ConnInfo each websocket connection
type ConnInfo struct {
	wsConn    *websocket.Conn
	name      string
	msgLocker sync.Mutex
}

// edgeConnMgr global connection map
var edgeConnMgr = map[string]*ConnInfo{}
var connLocker sync.Mutex

const (
	defaultReadLimitInBytes  = 1 * constants.MB
	defaultHeartbeatDuration = 5 * time.Second
	defaultHeartbeatTimeout  = 15 * time.Second
	heartbeatCheckInterval   = 5 * time.Second
)

// RegistryConn add one entry to global connection map
func RegistryConn(name string, ws *websocket.Conn) error {
	connLocker.Lock()
	defer connLocker.Unlock()
	if _, ok := edgeConnMgr[name]; ok {
		return fmt.Errorf("%v is already connected, new connection request is refused", name)
	}
	edgeConnMgr[name] = &ConnInfo{
		wsConn: ws,
		name:   name,
	}
	return nil
}

// UnRegistryConn remove one entry to global connection map
func UnRegistryConn(name string) error {
	connLocker.Lock()
	defer connLocker.Unlock()
	delete(edgeConnMgr, name)
	return nil
}

// GetConnByModName find an active ws conn by global connection map
func GetConnByModName(name string) (*ConnInfo, error) {
	connLocker.Lock()
	defer connLocker.Unlock()
	conn, ok := edgeConnMgr[name]
	if !ok {
		return nil, fmt.Errorf("connection not found. client %v not connected", name)
	}
	return conn, nil
}

// SendMsgToWs send msg to ws conn
func SendMsgToWs(msg *model.Message, srcMod string) error {
	var (
		dataBytes []byte
		err       error
	)
	if srcMod == constants.ModDeviceOm {
		common.MsgResultOptLog(msg)
		dataBytes, err = common.MarshalKubeedgeMessage(msg)
	} else {
		dataBytes, err = json.Marshal(msg)
	}
	if err != nil {
		return fmt.Errorf("marshal new data error: %v", err)
	}

	conn, err := GetConnByModName(srcMod)
	if err != nil {
		return err
	}
	if conn == nil {
		return fmt.Errorf("invalid websocket connection")
	}
	conn.msgLocker.Lock()
	defer conn.msgLocker.Unlock()

	return conn.wsConn.WriteMessage(websocket.TextMessage, dataBytes)
}

// SendHeartbeatToPeer send ping or pong message to peer with write lock
func SendHeartbeatToPeer(msgType int, data string, srcMod string) error {
	conn, err := GetConnByModName(srcMod)
	if err != nil {
		return err
	}
	if conn == nil {
		return fmt.Errorf("invalid websocket connection")
	}
	conn.msgLocker.Lock()
	defer conn.msgLocker.Unlock()
	switch msgType {
	case websocket.PingMessage, websocket.PongMessage:
		if err = conn.wsConn.WriteMessage(msgType, []byte(data)); err != nil {
			return fmt.Errorf("write heartbeat to peer failed: %v", err)
		}
	default:
		return fmt.Errorf("invalid websocket heartbeat message type: %v", msgType)
	}
	return nil
}
