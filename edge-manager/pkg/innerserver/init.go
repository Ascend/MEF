// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package innerserver module manager the inner websocket server
package innerserver

import (
	"context"
	"sync"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/constants"
)

// WsInnerServer wraps the struct WebSocketServer
type WsInnerServer struct {
	wsPort    int
	writeLock sync.RWMutex
	ctx       context.Context
	enable    bool
}

var server WsInnerServer

// NewInnerServer new cloud server
func NewInnerServer(enable bool, wsPort int) *WsInnerServer {
	server = WsInnerServer{
		wsPort: wsPort,
		ctx:    context.Background(),
		enable: enable,
	}
	return &server
}

// Name returns the name of websocket connection module
func (c *WsInnerServer) Name() string {
	return common.InnerServerName
}

// Enable indicates whether this module is enabled
func (c *WsInnerServer) Enable() bool {
	return c.enable
}

// Start initializes the websocket server
func (c *WsInnerServer) Start() {
	for count := 0; count < constants.ServerInitRetryCount; count++ {
		c.start()
		hwlog.RunLog.Error("init websocket server failed. Restart server later")
		time.Sleep(constants.ServerInitRetryInterval)
	}
	hwlog.RunLog.Error("init websocket server failed after maximum number of retry")
}

func (c *WsInnerServer) start() {
	hwlog.RunLog.Info("----------------inner server start----------------")
	if err := InitServer(); err != nil {
		hwlog.RunLog.Errorf("init websocket server failed: %v", err)
		return
	}

	hwlog.RunLog.Info("init websocket server success")
	for {
		select {
		case _, ok := <-c.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("has listened stop signal")
			return
		default:
		}

		message, err := modulemgr.ReceiveMessage(c.Name())
		if err != nil {
			hwlog.RunLog.Errorf("module [%s] receive message from channel failed, error: %v", c.Name(), err)
			continue
		}

		go c.dispatch(message)
	}
}

func (c *WsInnerServer) dispatch(message *model.Message) {
	sender := getWsSender()

	if err := sender.Send(message.GetNodeId(), message); err != nil {
		hwlog.RunLog.Errorf("send ws msg to %s failed: %s", message.GetNodeId(), err.Error())
	}
}
