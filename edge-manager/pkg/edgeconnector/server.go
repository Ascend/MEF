// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeconnector the websocket server config
package edgeconnector

import (
	"edge-manager/pkg/edgeconnector/common"
	"fmt"
	"github.com/gorilla/websocket"
	"huawei.com/mindx/common/hwlog"
	"net/http"
	"time"
)

// WebSocketServer defines the websocket server config
type WebSocketServer struct {
	server        *http.Server
	WriteDeadline time.Duration
	ReadDeadline  time.Duration
	allClients    map[string]*websocket.Conn
	isConnMap     map[string]bool
}

// Online indicates edge-installer is online, Offline indicates edge-installer is offline
const (
	Online  = true
	Offline = false
)

// NewWebsocketServer instantiates WebSocketServer struct
func NewWebsocketServer() (*WebSocketServer, error) {
	s := WebSocketServer{
		server: &http.Server{
			Addr: fmt.Sprintf("%s:%s", Config.ServerAddress, Config.Port),
		},
		WriteDeadline: common.WriteDeadline,
		ReadDeadline:  common.ReadDeadline,
		allClients:    make(map[string]*websocket.Conn),
		isConnMap:     make(map[string]bool),
	}
	return &s, nil
}

// StartWebsocketServer starts a websocket server
func (w *WebSocketServer) StartWebsocketServer() {
	http.HandleFunc("/", w.ServeHTTP)
	hwlog.RunLog.Info("websocket server is listening....")
	if err := w.server.ListenAndServe(); err != nil {
		hwlog.RunLog.Error("error during websocket server listening: ", err)
		return
	}
}

func (w *WebSocketServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  common.ReadBufferSize,
		WriteBufferSize: common.WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		hwlog.RunLog.Error("error during connection upgrade: ", err)
		return
	}
	hwlog.RunLog.Info("websocket connection is ok")
	nodeID := req.Header.Get("SerialNumber")
	w.allClients[nodeID] = conn
	w.notifyTasks(nodeID)
}

func (w *WebSocketServer) notifyTasks(nodeID string) {
	w.SetClientStatus(nodeID, Online)
}

// IsClientConnected indicates whether the edge-installer is connected
func (w *WebSocketServer) IsClientConnected(nodeID string) bool {
	return w.isConnMap[nodeID]
}

// CloseConnection closes the websocket connection
func (w *WebSocketServer) CloseConnection(nodeID string) {
	w.SetClientStatus(nodeID, Offline)
	if err := w.server.Close(); err != nil {
		hwlog.RunLog.Error("close websocket connection error: ", err)
	}
}

// SetClientStatus sets the edge-installer status
func (w *WebSocketServer) SetClientStatus(nodeID string, connectStatus bool) {
	w.isConnMap[nodeID] = connectStatus
}
