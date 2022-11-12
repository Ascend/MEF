// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"context"
	"edge-manager/module_manager"
	"edge-manager/module_manager/model"
	"edge-manager/pkg/common"
	"edge-manager/pkg/database"
	"fmt"

	"huawei.com/mindx/common/hwlog"
)

type handlerFunc func(req interface{}) common.RespMsg

type nodeManager struct {
	enable bool
	ctx    context.Context
}

// NewNodeManager new node manager
func NewNodeManager(enable bool) *nodeManager {
	nm := &nodeManager{
		enable: enable,
		ctx:    context.Background(),
	}
	return nm
}

func (node *nodeManager) Name() string {
	return common.NodeManagerName
}

func (node *nodeManager) Enable() bool {
	return node.enable
}

func (node *nodeManager) Start() {
	if err := database.CreateTableIfNotExists(NodeGroup{}); err != nil {
		hwlog.RunLog.Error("create node group database table failed")
		return
	}
	if err := database.CreateTableIfNotExists(NodeInfo{}); err != nil {
		hwlog.RunLog.Error("create node database table failed")
		return
	}
	for {
		select {
		case _, ok := <-node.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("has listened stop signal")
			return
		default:
		}
		req, err := module_manager.ReceiveMessage(common.NodeManagerName)
		hwlog.RunLog.Debugf("%s revice requst from restful service", common.NodeManagerName)
		if err != nil {
			hwlog.RunLog.Errorf("%s revice requst from restful service failed", common.NodeManagerName)
			continue
		}
		msg := methodSelect(req)
		if msg == nil {
			hwlog.RunLog.Error("%s get method by option and resource failed", common.NodeManagerName)
			continue
		}
		resp, err := req.NewResponse()
		if err != nil {
			hwlog.RunLog.Error("%s new response failed", common.NodeManagerName)
			continue
		}
		resp.FillContent(msg)
		if err = module_manager.SendMessage(resp); err != nil {
			hwlog.RunLog.Error("%s send response failed", common.NodeManagerName)
			continue
		}
	}
}

func methodSelect(req *model.Message) *common.RespMsg {
	var res common.RespMsg
	method, exit := nodeMethodList()[combine(req.GetOption(), req.GetResource())]
	if !exit {
		return nil
	}
	res = method(req.GetContent())
	return &res
}

func nodeMethodList() map[string]handlerFunc {
	return map[string]handlerFunc{
		combine(common.Create, common.Node): CreateNode,
	}
}

func combine(option, resource string) string {
	return fmt.Sprintf("%s%s", option, resource)
}
