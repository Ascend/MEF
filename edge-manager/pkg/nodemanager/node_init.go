// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"

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
func NewNodeManager(enable bool, ctx context.Context) *nodeManager {
	nm := &nodeManager{
		enable: enable,
		ctx:    ctx,
	}
	return nm
}

func (node *nodeManager) Name() string {
	return common.NodeManagerName
}

func (node *nodeManager) Enable() bool {
	if node.enable {
		if err := initNodeTable(); err != nil {
			hwlog.RunLog.Errorf("module (%s) init database table failed, cannot enable", node.Name())
			return !node.enable
		}
		if err := initNodeSyncService(); err != nil {
			hwlog.RunLog.Errorf("module (%s) init node status service failed, cannot enable", node.Name())
			return !node.enable
		}
	}
	return node.enable
}

func (node *nodeManager) Start() {
	hwlog.RunLog.Info("----------------node manager start----------------")
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
		req, err := modulemanager.ReceiveMessage(node.Name())
		hwlog.RunLog.Debugf("%s receive request from restful service", node.Name())
		if err != nil {
			hwlog.RunLog.Errorf("%s receive request from restful service failed", node.Name())
			continue
		}
		msg, err := dispatchMsg(req)
		if err != nil {
			hwlog.RunLog.Errorf("%s get method by option and resource failed", node.Name())
			continue
		}

		if !req.GetIsSync() {
			continue
		}

		resp, err := req.NewResponse()
		if err != nil {
			hwlog.RunLog.Errorf("%s new response failed: %v", node.Name(), err)
			continue
		}
		resp.FillContent(msg)
		if err = modulemanager.SendMessage(resp); err != nil {
			hwlog.RunLog.Errorf("%s send response failed: %v", node.Name(), err)
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

func dispatchMsg(req *model.Message) (*common.RespMsg, error) {
	var res common.RespMsg
	method, exit := handlerFuncMap[common.Combine(req.GetOption(), req.GetResource())]
	if !exit {
		return nil, fmt.Errorf("method not found for router: option=%s, resource=%s", req.GetOption(), req.GetResource())
	}
	res = method(req.GetContent())
	return &res, nil
}

var (
	nodeUrlRootPath   = "/edgemanager/v1/node"
	nodeGroupRootPath = "/edgemanager/v1/nodegroup"
)

var handlerFuncMap = map[string]handlerFunc{
	common.Combine(http.MethodGet, filepath.Join(nodeUrlRootPath, "stats")):          getNodeStatistics,
	common.Combine(http.MethodGet, nodeUrlRootPath):                                  getNodeDetail,
	common.Combine(http.MethodPatch, nodeUrlRootPath):                                modifyNode,
	common.Combine(http.MethodPost, filepath.Join(nodeUrlRootPath, "batch-delete")):  batchDeleteNode,
	common.Combine(http.MethodGet, filepath.Join(nodeUrlRootPath, "list/managed")):   listManagedNode,
	common.Combine(http.MethodGet, filepath.Join(nodeUrlRootPath, "list/unmanaged")): listUnmanagedNode,
	common.Combine(http.MethodGet, filepath.Join(nodeUrlRootPath, "list")):           listNode,
	common.Combine(http.MethodPost, filepath.Join(nodeUrlRootPath, "add")):           addUnManagedNode,
	common.Combine(http.MethodGet, filepath.Join(nodeUrlRootPath, "capability")):     getNodesCapability,

	common.Combine(http.MethodPost, nodeGroupRootPath):                                     createGroup,
	common.Combine(http.MethodGet, filepath.Join(nodeGroupRootPath, "stats")):              getGroupNodeStatistics,
	common.Combine(http.MethodGet, nodeGroupRootPath):                                      getEdgeNodeGroupDetail,
	common.Combine(http.MethodPatch, nodeGroupRootPath):                                    modifyNodeGroup,
	common.Combine(http.MethodPost, filepath.Join(nodeGroupRootPath, "batch-delete")):      batchDeleteNodeGroup,
	common.Combine(http.MethodGet, filepath.Join(nodeGroupRootPath, "list")):               listEdgeNodeGroup,
	common.Combine(http.MethodPost, filepath.Join(nodeGroupRootPath, "node")):              addNodeRelation,
	common.Combine(http.MethodPost, filepath.Join(nodeGroupRootPath, "node/batch-delete")): deleteNodeFromGroup,
	common.Combine(http.MethodPost, filepath.Join(nodeGroupRootPath, "pod/batch-delete")):  batchDeleteNodeRelation,

	common.Combine(common.Inner, common.Node):                innerGetNodeInfoByUniqueName,
	common.Combine(common.Inner, common.NodeGroup):           innerGetNodeGroupInfosByIds,
	common.Combine(common.Inner, common.NodeSoftwareInfo):    innerGetNodeSoftwareInfo,
	common.Combine(common.Inner, common.NodeStatus):          innerGetNodeStatus,
	common.Combine(common.Inner, common.CheckResource):       innerCheckNodeGroupResReq,
	common.Combine(common.Inner, common.UpdateResource):      innerUpdateNodeGroupResReq,
	common.Combine(common.OptReport, common.ResSoftwareInfo): updateNodeSoftwareInfo,
	common.Combine(common.Inner, common.NodeList):            innerAllNodeInfos,
	common.Combine(common.Inner, common.NodeID):              innerGetNodesByNodeGroupID,
}
