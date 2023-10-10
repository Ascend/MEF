// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-manager/pkg/constants"

	"huawei.com/mindxedge/base/common"
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

		req, err := modulemgr.ReceiveMessage(node.Name())
		hwlog.RunLog.Debugf("%s receive request from restful service", node.Name())
		if err != nil {
			hwlog.RunLog.Errorf("%s receive request from restful service failed", node.Name())
			continue
		}

		go node.dispatch(req)
	}
}

func (node *nodeManager) dispatch(req *model.Message) {
	var err error
	msg, err := selectMethod(req)
	if err != nil {
		msg, err = methodWithOpLogSelect(handlerWithOpLogFuncMap, req)
		if err != nil {
			hwlog.RunLog.Error(err)
			return
		}
	}
	if !req.GetIsSync() {
		return
	}

	resp, err := req.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("%s new response failed: %v", node.Name(), err)
		return
	}
	resp.FillContent(msg)
	if err = modulemgr.SendMessage(resp); err != nil {
		hwlog.RunLog.Errorf("%s send response failed: %v", node.Name(), err)
		return
	}
}

func initNodeTable() error {
	if err := database.CreateTableIfNotExist(NodeGroup{}); err != nil {
		hwlog.RunLog.Error("create node group database table failed")
		return err
	}
	if err := database.CreateTableIfNotExist(NodeInfo{}); err != nil {
		hwlog.RunLog.Error("create node database table failed")
		return err
	}
	if err := database.CreateTableIfNotExist(NodeRelation{}); err != nil {
		hwlog.RunLog.Error("create node database table failed")
		return err
	}
	return nil
}

func selectMethod(req *model.Message) (*common.RespMsg, error) {
	var res common.RespMsg
	method, exit := handlerFuncMap[common.Combine(req.GetOption(), req.GetResource())]
	if !exit {
		return nil, fmt.Errorf("method not found for router: option=%s, resource=%s", req.GetOption(), req.GetResource())
	}
	res = method(req.GetContent())
	return &res, nil
}

func methodWithOpLogSelect(funcMap map[string]handlerFunc, req *model.Message) (*common.RespMsg, error) {
	sn := req.GetNodeId()
	ip := req.GetIp()
	var res common.RespMsg
	method, ok := funcMap[common.Combine(req.GetOption(), req.GetResource())]
	if !ok {
		return nil, fmt.Errorf("handler func is not exist, option: %s, resource: %s", req.GetOption(),
			req.GetResource())
	}
	hwlog.RunLog.Infof("%v [%v:%v] %v %v start",
		time.Now().Format(time.RFC3339Nano), ip, sn, req.GetOption(), req.GetResource())
	res = method(req.GetContent())
	if res.Status == common.Success {
		hwlog.RunLog.Infof("%v [%v:%v] %v %v success",
			time.Now().Format(time.RFC3339Nano), ip, sn, req.GetOption(), req.GetResource())
	} else {
		hwlog.RunLog.Errorf("%v [%v:%v] %v %v failed",
			time.Now().Format(time.RFC3339Nano), ip, sn, req.GetOption(), req.GetResource())
	}
	return &res, nil
}

var (
	nodeUrlRootPath   = "/edgemanager/v1/node"
	nodeGroupRootPath = "/edgemanager/v1/nodegroup"
)

var handlerFuncMap = map[string]handlerFunc{
	common.Combine(http.MethodGet, filepath.Join(nodeUrlRootPath, "stats")):                   getNodeStatistics,
	common.Combine(http.MethodGet, nodeUrlRootPath):                                           getNodeDetail,
	common.Combine(http.MethodPatch, nodeUrlRootPath):                                         modifyNode,
	common.Combine(http.MethodPost, filepath.Join(nodeUrlRootPath, "batch-delete")):           batchDeleteNode,
	common.Combine(http.MethodPost, filepath.Join(nodeUrlRootPath, "batch-delete/unmanaged")): deleteUnManagedNode,
	common.Combine(http.MethodGet, filepath.Join(nodeUrlRootPath, "list/managed")):            listManagedNode,
	common.Combine(http.MethodGet, filepath.Join(nodeUrlRootPath, "list/unmanaged")):          listUnmanagedNode,
	common.Combine(http.MethodGet, filepath.Join(nodeUrlRootPath, "list")):                    listNode,
	common.Combine(http.MethodPost, filepath.Join(nodeUrlRootPath, "add")):                    addUnManagedNode,

	common.Combine(http.MethodPost, nodeGroupRootPath):                                     createNodeGroup,
	common.Combine(http.MethodPatch, nodeGroupRootPath):                                    modifyNodeGroup,
	common.Combine(http.MethodGet, nodeGroupRootPath):                                      getNodeGroupDetail,
	common.Combine(http.MethodGet, filepath.Join(nodeGroupRootPath, "stats")):              getNodeGroupStatistics,
	common.Combine(http.MethodGet, filepath.Join(nodeGroupRootPath, "list")):               listNodeGroup,
	common.Combine(http.MethodPost, filepath.Join(nodeGroupRootPath, "node")):              addNodeRelation,
	common.Combine(http.MethodPost, filepath.Join(nodeGroupRootPath, "batch-delete")):      batchDeleteNodeGroup,
	common.Combine(http.MethodPost, filepath.Join(nodeGroupRootPath, "node/batch-delete")): deleteNodeFromGroup,
	common.Combine(http.MethodPost, filepath.Join(nodeGroupRootPath, "pod/batch-delete")):  batchDeleteNodeRelation,

	common.Combine(common.Inner, common.Node):                innerGetNodeInfoByUniqueName,
	common.Combine(common.Inner, constants.NodeSerialNumber): innerGetNodeUniqueNameByID,
	common.Combine(common.Inner, common.NodeGroup):           innerGetNodeGroupInfosByIds,
	common.Combine(common.Inner, common.NodeSoftwareInfo):    innerGetNodeSoftwareInfo,
	common.Combine(common.Inner, common.NodeStatus):          innerGetNodeStatus,
	common.Combine(common.Inner, common.CheckResource):       innerCheckNodeGroupResReq,
	common.Combine(common.Inner, common.UpdateResource):      innerUpdateNodeGroupResReq,
	common.Combine(common.Inner, common.NodeList):            innerAllNodeInfos,
	common.Combine(common.Inner, common.NodeID):              innerGetNodesByNodeGroupID,
	common.Combine(common.Get, common.GetIpBySn):             innerGetIpBySn,
	common.Combine(common.Get, common.GetSnsByGroup):         innerGetNodeSnsByGroupId,
}

var handlerWithOpLogFuncMap = map[string]handlerFunc{
	common.Combine(common.OptReport, common.ResSoftwareInfo): updateNodeSoftwareInfo,
}
