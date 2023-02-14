// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package cloudhub module edge-connector init
package cloudhub

import (
	"context"
	"sync"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

// CloudServer wraps the struct WebSocketServer
type CloudServer struct {
	wsPort    int
	authPort  int
	writeLock sync.RWMutex
	ctx       context.Context
	enable    bool
}

var server CloudServer

// NewCloudServer new cloud server
func NewCloudServer(enable bool, wsPort, authPort int) *CloudServer {
	server = CloudServer{
		wsPort:   wsPort,
		authPort: authPort,
		ctx:      context.Background(),
		enable:   enable,
	}
	return &server
}

// Name returns the name of websocket connection module
func (c *CloudServer) Name() string {
	return common.CloudHubName
}

// Enable indicates whether this module is enabled
func (c *CloudServer) Enable() bool {
	return c.enable
}

// Start initializes the websocket server
func (c *CloudServer) Start() {
	hwlog.RunLog.Info("----------------cloud hub start----------------")
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

func (c *CloudServer) sendToClient(msg *model.Message) {
	sender, err := GetSvrSender()
	if err != nil {
		hwlog.RunLog.Errorf("send to client [%s] failed", msg.GetNodeId())
		return
	}
	if err = sender.Send(msg.GetNodeId(), msg); err != nil {
		hwlog.RunLog.Errorf("cloud hub send msg to edge node error: %v, operation is [%s], resource is [%s]",
			err, msg.GetOption(), msg.GetResource())
		return
	}

	hwlog.RunLog.Infof("cloud hub send msg to edge node success, operation is [%s], resource is [%s]",
		msg.GetOption(), msg.GetResource())
}
