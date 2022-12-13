// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"context"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"

	"edge-manager/pkg/database"
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
	if node.enable {
		if err := initNodeTable(); err != nil {
			hwlog.RunLog.Errorf("module (%s) init database table failed, cannot enable", common.NodeManagerName)
			return !node.enable
		}
		if err := initNodeStatusService(); err != nil {
			hwlog.RunLog.Errorf("module (%s) init node status service failed, cannot enable", common.NodeManagerName)
			return !node.enable
		}
	}
	return node.enable
}

func (node *nodeManager) Start() {
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
		req, err := modulemanager.ReceiveMessage(common.NodeManagerName)
		hwlog.RunLog.Debugf("%s revice requst from restful service", common.NodeManagerName)
		if err != nil {
			hwlog.RunLog.Errorf("%s revice requst from restful service failed", common.NodeManagerName)
			continue
		}
		msg := methodSelect(req)
		if msg == nil {
			hwlog.RunLog.Errorf("%s get method by option and resource failed", common.NodeManagerName)
			continue
		}
		resp, err := req.NewResponse()
		if err != nil {
			hwlog.RunLog.Errorf("%s new response failed", common.NodeManagerName)
			continue
		}
		resp.FillContent(msg)
		if err = modulemanager.SendMessage(resp); err != nil {
			hwlog.RunLog.Errorf("%s send response failed", common.NodeManagerName)
			continue
		}
	}
}

func initNodeTable() error {
	if err := database.CreateTableIfNotExists(NodeGroup{}); err != nil {
		hwlog.RunLog.Error("create node group database table failed")
		return err
	}
	if err := database.CreateTableIfNotExists(NodeInfo{}); err != nil {
		hwlog.RunLog.Error("create node database table failed")
		return err
	}
	if err := database.CreateTableIfNotExists(NodeRelation{}); err != nil {
		hwlog.RunLog.Error("create node database table failed")
		return err
	}
	return nil
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
		combine(common.Create, common.Node):         createNode,
		combine(common.Create, common.NodeGroup):    createGroup,
		combine(common.List, common.Node):           listNode,
		combine(common.List, common.NodeUnManaged):  listNodeUnManaged,
		combine(common.Get, common.Node):            getNodeDetail,
		combine(common.Update, common.Node):         modifyNode,
		combine(common.Delete, common.Node):         batchDeleteNode,
		combine(common.Add, common.Node):            addUnManagedNode,
		combine(common.Add, common.NodeRelation):    addNodeRelation,
		combine(common.Delete, common.NodeRelation): batchDeleteNodeRelation,
		combine(common.Get, common.NodeStatistics):  getNodeStatistics,
		combine(common.List, common.NodeGroup):      listEdgeNodeGroup,
		combine(common.Get, common.NodeGroup):       getEdgeNodeGroupDetail,
		combine(common.Delete, common.NodeGroup):    batchDeleteNodeGroup,

		combine(common.Inner, common.Node):       innerGetNodeInfoByName,
		combine(common.Inner, common.NodeGroup):  innerGetNodeGroupInfoById,
		combine(common.Inner, common.NodeStatus): innerGetNodeStatus,
	}
}

func combine(option, resource string) string {
	return fmt.Sprintf("%s%s", option, resource)
}
