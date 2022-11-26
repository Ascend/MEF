// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeconnector the websocket server config
package edgeconnector

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"huawei.com/mindxedge/base/modulemanager/model"

	"github.com/gorilla/websocket"
	"huawei.com/mindx/common/hwlog"
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

func newWebsocketServer() *WebSocketServer {
	return &WebSocketServer{
		server: &http.Server{
			Addr: fmt.Sprintf("%s:%s", Config.ServerAddress, Config.Port),
		},
		WriteDeadline: WriteDeadline,
		ReadDeadline:  ReadDeadline,
		allClients:    make(map[string]*websocket.Conn),
		isConnMap:     make(map[string]bool),
	}
}

func (w *WebSocketServer) startWebsocketServer() {
	http.HandleFunc("/", w.ServeHTTP)
	hwlog.RunLog.Info("websocket server is listening...")
	if err := w.server.ListenAndServe(); err != nil {
		hwlog.RunLog.Errorf("error during websocket server listening: %v", err)
		return
	}
}

func (w *WebSocketServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  ReadBufferSize,
		WriteBufferSize: WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		hwlog.RunLog.Errorf("error during connection upgrade: %v", err)
		return
	}

	hwlog.RunLog.Info("websocket connection is ok")
	nodeID := req.Header.Get("SerialNumber")
	w.allClients[nodeID] = conn
	w.notifyTasks(nodeID)
}

// Send sends message to edge-installer
func (w *WebSocketServer) Send(message *model.Message, nodeId string) error {
	conn := w.allClients[nodeId]

	if err := conn.SetWriteDeadline(time.Now().Add(WriteCenterDeadline)); err != nil {
		hwlog.RunLog.Error("write message time out")
		return err
	}

	var err error
	for i := 0; i < WriteRetryCount; i++ {
		err = conn.WriteJSON(message)
		if err == nil {
			return nil
		}
	}

	return errors.New("max retry count to send to edge-installer")
}

// Receive receives message from edge-installer
func (w *WebSocketServer) Receive(conn *websocket.Conn) (*model.Message, error) {
	var message *model.Message
	_, r, err := conn.NextReader()
	if err != nil {
		hwlog.RunLog.Errorf("read error is %v", err)
		return message, err
	}
	err = json.NewDecoder(r).Decode(&message)
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	if err != nil {
		hwlog.RunLog.Errorf("read json failed, error: %v", err)
		return message, err
	}
	return message, nil
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
		hwlog.RunLog.Errorf("close websocket connection error: %v", err)
	}
}

// SetClientStatus sets the edge-installer status
func (w *WebSocketServer) SetClientStatus(nodeID string, connectStatus bool) {
	w.isConnMap[nodeID] = connectStatus
}
