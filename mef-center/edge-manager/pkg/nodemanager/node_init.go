// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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

type handlerFunc func(msg *model.Message) common.RespMsg

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
	if err = resp.FillContent(msg); err != nil {
		hwlog.RunLog.Errorf("%s fill content failed: %v", node.Name(), err)
		return
	}
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
	res = method(req)
	return &res, nil
}

func methodWithOpLogSelect(funcMap map[string]handlerFunc, req *model.Message) (*common.RespMsg, error) {
	sn := req.GetPeerInfo().Sn
	ip := req.GetPeerInfo().Ip
	var res common.RespMsg
	method, ok := funcMap[common.Combine(req.GetOption(), req.GetResource())]
	if !ok {
		return nil, fmt.Errorf("handler func is not exist, option: %s, resource: %s", req.GetOption(),
			req.GetResource())
	}
	hwlog.RunLog.Infof("%v [%v:%v] %v %v start",
		time.Now().Format(time.RFC3339Nano), ip, sn, req.GetOption(), req.GetResource())
	res = method(req)
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

	common.Combine(common.Inner, common.Node):             innerGetNodeInfoByUniqueName,
	common.Combine(common.Inner, constants.NodeSnAndIp):   innerGetNodeSnAndIpByID,
	common.Combine(common.Inner, common.NodeGroup):        innerGetNodeGroupInfosByIds,
	common.Combine(common.Inner, common.NodeSoftwareInfo): innerGetNodeSoftwareInfo,
	common.Combine(common.Inner, common.NodeStatus):       innerGetNodeStatus,
	common.Combine(common.Inner, common.CheckResource):    innerCheckNodeGroupResReq,
	common.Combine(common.Inner, common.UpdateResource):   innerUpdateNodeGroupResReq,
	common.Combine(common.Inner, common.NodeList):         innerAllNodeInfos,
	common.Combine(common.Inner, common.NodeID):           innerGetNodesByNodeGroupID,
	common.Combine(common.Get, common.GetIpBySn):          innerGetIpBySn,
	common.Combine(common.Get, common.GetSnsByGroup):      innerGetNodeSnsByGroupId,
}

var handlerWithOpLogFuncMap = map[string]handlerFunc{
	common.Combine(common.OptReport, common.ResSoftwareInfo): updateNodeSoftwareInfo,
}
