// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeconnector module edge-connector init
package edgeconnector

import (
	"context"
	"sync"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

// Connector wraps the struct WebSocketServer
type Connector struct {
	wsPort    int
	writeLock sync.RWMutex
	ctx       context.Context
	enable    bool
}

var connector Connector

// NewConnector new connector
func NewConnector(enable bool, wsPort int) *Connector {
	connector = Connector{
		wsPort: wsPort,
		ctx:    context.Background(),
		enable: enable,
	}
	return &connector
}

// Name returns the name of websocket connection module
func (c *Connector) Name() string {
	return common.EdgeConnectorName
}

// Enable indicates whether this module is enabled
func (c *Connector) Enable() bool {
	return c.enable
}

// Start initializes the websocket server
func (c *Connector) Start() {
	hwlog.RunLog.Info("----------------edge connector start----------------")
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

		message, err := modulemanager.ReceiveMessage(c.Name())
		if err != nil {
			hwlog.RunLog.Errorf("module [%s] receive message from channel failed, error: %v", c.Name(), err)
			continue
		}
		c.sendToClient(message)
	}
}

func (c *Connector) sendToClient(msg *model.Message) {
	sender := GetSvrSender()
	if err := sender.Send(msg.GetNodeId(), msg); err != nil {
		hwlog.RunLog.Errorf("edge-connector send msg to edge node error: %v, operation is [%s], resource is [%s]",
			err, msg.GetOption(), msg.GetResource())
		return
	}

	hwlog.RunLog.Infof("edge-connector send msg to edge node success, operation is [%s], resource is [%s]",
		msg.GetOption(), msg.GetResource())
	return
}
