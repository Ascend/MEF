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

	"github.com/gin-gonic/gin"

	"huawei.com/mindxedge/base/modulemanager/model"

	"github.com/gorilla/websocket"
	"huawei.com/mindx/common/hwlog"
)

// WebSocketServer defines the websocket server config
type WebSocketServer struct {
	server        *gin.Engine
	WriteDeadline time.Duration
	ReadDeadline  time.Duration
	allClients    map[string]*websocket.Conn
	isConnMap     map[string]bool
}

func newWebsocketServer(c *gin.Engine) *WebSocketServer {
	return &WebSocketServer{
		server:        c,
		WriteDeadline: WriteDeadline,
		ReadDeadline:  ReadDeadline,
		allClients:    make(map[string]*websocket.Conn),
		isConnMap:     make(map[string]bool),
	}
}

func (w *WebSocketServer) startWebsocketServer() {
	w.server.GET("/", w.ServeHTTP)
	hwlog.RunLog.Info("websocket server is listening...")

	if err := w.server.Run(fmt.Sprintf(":%s", Config.Port)); err != nil {
		hwlog.RunLog.Errorf("error during websocket server listening: %v", err)
		return
	}

	return
}

func (w *WebSocketServer) ServeHTTP(c *gin.Context) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  ReadBufferSize,
		WriteBufferSize: WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		hwlog.RunLog.Errorf("error during connection upgrade: %v", err)
		return
	}

	nodeID := c.Request.Header.Get("SerialNumber")
	hwlog.RunLog.Infof("websocket connection with [%s] is ok", nodeID)
	w.allClients[nodeID] = conn
	w.notifyTasks(nodeID)
}

// Send sends message to edge-installer
func (w *WebSocketServer) Send(message *model.Message, nodeId string) error {
	conn := w.allClients[nodeId]
	if conn == nil {
		return errors.New("conn is nil")
	}

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
	if conn == nil {
		return nil, errors.New("conn is nil")
	}

	message, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Error("new message failed")
		return message, err
	}

	_, r, err := conn.NextReader()
	if err != nil {
		hwlog.RunLog.Errorf("read error is %v", err)
		return message, err
	}

	err = json.NewDecoder(r).Decode(message)
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	if err != nil {
		hwlog.RunLog.Errorf("read json failed, error: %v", err)
		return message, err
	}

	return message, nil
}

func (w *WebSocketServer) notifyTasks(nodeId string) {
	w.SetClientStatus(nodeId, Online)
}

// IsClientConnected indicates whether the edge-installer is connected
func (w *WebSocketServer) IsClientConnected(nodeId string) bool {
	return w.isConnMap[nodeId]
}

// CloseConnection closes the websocket connection
func (w *WebSocketServer) CloseConnection(nodeId string) {
	w.SetClientStatus(nodeId, Offline)
	if err := w.server.DELETE(fmt.Sprintf("%s:%s", Config.ServerAddress, Config.Port)); err != nil {
		hwlog.RunLog.Errorf("close websocket connection error: %v", err)
	}
}

// SetClientStatus sets the edge-installer status
func (w *WebSocketServer) SetClientStatus(nodeId string, connectStatus bool) {
	w.isConnMap[nodeId] = connectStatus
}
