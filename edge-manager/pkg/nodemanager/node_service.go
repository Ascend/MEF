// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"edge-manager/pkg/util"
	"fmt"
	"strings"
	"time"

	"huawei.com/mindxedge/base/common"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/kubeclient"
)

// CreateNode Create Node
func createNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start create node")
	var req util.CreateEdgeNodeReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	total, err := GetTableCount(NodeInfo{})
	if err != nil {
		hwlog.RunLog.Error("get node table num failed")
		return common.RespMsg{Status: "", Msg: "get node table num failed", Data: nil}
	}
	if total >= maxNode {
		hwlog.RunLog.Error("node number is enough, connot create")
		return common.RespMsg{Status: "", Msg: "node number is enough, connot create", Data: nil}
	}
	node := &NodeInfo{
		Description: req.Description,
		UniqueName:  req.UniqueName,
		NodeName:    req.NodeName,
		Status:      statusOffline,
		IsManaged:   true,
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdateAt:    time.Now().Format(TimeFormat),
	}
	if err = NodeServiceInstance().createNode(node); err != nil {
		if strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
			hwlog.RunLog.Error("node name is duplicate")
			return common.RespMsg{Status: "", Msg: "node name is duplicate", Data: nil}
		}
		hwlog.RunLog.Error("node db create failed")
		return common.RespMsg{Status: "", Msg: "db create failed", Data: nil}
	}
	hwlog.RunLog.Info("node db create success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

// getNodeDetail get node detail
func getNodeDetail(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start get node detail")
	var req util.GetNodeDetailReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Error("get node detail convert request error")
		return common.RespMsg{Status: "", Msg: "convert request error", Data: nil}
	}
	if err := req.Check(); err != nil {
		hwlog.RunLog.Error("modify node check parameters failed")
		return common.RespMsg{Status: "", Msg: "check parameters failed", Data: nil}
	}
	nodeInfo, err := NodeServiceInstance().getNodeByID(req.Id)
	if err != nil {
		hwlog.RunLog.Error("get node detail db query error")
		return common.RespMsg{Status: "", Msg: "db query error", Data: nil}
	}
	nodeRelation, err := NodeServiceInstance().getNodeRelationByNodeId(req.Id)
	if err != nil {
		hwlog.RunLog.Error("get node detail db query error")
		return common.RespMsg{Status: "", Msg: "db query error", Data: nil}
	}
	nodeGroup, err := NodeServiceInstance().getNodeGroupByID(nodeRelation.GroupID)
	if err != nil {
		hwlog.RunLog.Error("get node detail db query error")
		return common.RespMsg{Status: "", Msg: "db query error", Data: nil}
	}
	resp := util.GetNodeDetailResp{
		Id:          nodeInfo.ID,
		NodeName:    nodeInfo.NodeName,
		UniqueName:  nodeInfo.UniqueName,
		Description: nodeInfo.Description,
		Status:      nodeInfo.Status,
		CreatedAt:   nodeInfo.CreatedAt,
		UpdatedAt:   nodeInfo.UpdateAt,
		Cpu:         nodeInfo.CPUCore,
		Memory:      nodeInfo.Memory,
		Npu:         nodeInfo.NPUType,
		NodeType:    nodeInfo.NodeType,
		NodeGroup:   nodeGroup.GroupName,
	}
	hwlog.RunLog.Info("node detail db query success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

func modifyNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start modify node")
	var req util.ModifyNodeGroupReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Error("modify node convert request error")
		return common.RespMsg{Status: "", Msg: "convert request error", Data: nil}
	}
	if err := req.Check(); err != nil {
		hwlog.RunLog.Error("modify node check parameters failed")
		return common.RespMsg{Status: "", Msg: "check parameters failed", Data: nil}
	}
	updatedColumns := map[string]interface{}{
		"NodeName":    req.NodeName,
		"Description": req.Description,
	}
	err := NodeServiceInstance().updateNode(req.NodeId, updatedColumns)
	if err != nil {
		if strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
			hwlog.RunLog.Error("node name is duplicate")
			return common.RespMsg{Status: "", Msg: "node name is duplicate", Data: nil}
		}
		hwlog.RunLog.Error("modify node db update error")
		return common.RespMsg{Status: "", Msg: "db update error", Data: nil}
	}
	hwlog.RunLog.Info("modify node db update success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func getNodeStatistics(interface{}) common.RespMsg {
	hwlog.RunLog.Info("start get node statistics")
	resp := make(map[string]int64)
	allNodeStatus := []string{statusReady, statusNotReady, statusOffline, statusUnknown}
	for _, status := range allNodeStatus {
		nodeCount, err := NodeServiceInstance().countNodesByStatus(status)
		if err != nil {
			hwlog.RunLog.Error("get node statistics db query error")
			return common.RespMsg{Status: "", Msg: "db query error", Data: nil}
		}
		resp[status] = nodeCount
	}
	hwlog.RunLog.Info("get node statistics db query success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

// ListNode get node list
func listNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start list node managed")
	var req util.ListReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	nodes, err := NodeServiceInstance().listNodesByName(req.PageNum, req.PageSize, req.Name)
	if err == nil {
		hwlog.RunLog.Info("list node success")
		return common.RespMsg{Status: common.Success, Msg: "", Data: nodes}
	}
	if err == gorm.ErrRecordNotFound {
		hwlog.RunLog.Info("dont have any managed node")
		return common.RespMsg{Status: common.Success, Msg: "dont have any managed node", Data: nil}
	}
	hwlog.RunLog.Error("list node failed")
	return common.RespMsg{Status: "", Msg: "list node failed", Data: nil}
}

// ListNode get node list
func listNodeUnManaged(input interface{}) common.RespMsg {
	if err := autoAddUnmanagedNode(); err != nil {
		hwlog.RunLog.Error("auto add unmanaged node filed")
		return common.RespMsg{Status: "", Msg: "auto add unmanaged node filed", Data: nil}
	}
	hwlog.RunLog.Info("start list node unmanaged")
	var req util.ListReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	nodes, err := NodeServiceInstance().listUnManagedNodesByName(req.PageNum, req.PageSize, req.Name)
	if err == nil {
		hwlog.RunLog.Info("list node unmanaged success")
		return common.RespMsg{Status: common.Success, Msg: "", Data: nodes}
	}
	if err == gorm.ErrRecordNotFound {
		hwlog.RunLog.Info("dont have any unmanaged node")
		return common.RespMsg{Status: common.Success, Msg: "dont have any unmanaged node", Data: nil}
	}
	hwlog.RunLog.Error("list unmanaged node failed")
	return common.RespMsg{Status: "", Msg: "list unmanaged node failed", Data: nil}
}

func autoAddUnmanagedNode() error {
	realNodes, err := kubeclient.GetKubeClient().ListNode()
	if err != nil {
		return err
	}
	nodeDb, err := GetTableCount(NodeInfo{})
	if err != nil {
		hwlog.RunLog.Error("get node table num failed")
		return err
	}
	if len(realNodes.Items) == nodeDb {
		return nil
	}
	for _, node := range realNodes.Items {
		_, err := NodeServiceInstance().getNodeByUniqueName(node.Name)
		if err == nil {
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("get node by name(%s) failed", node.Name)
		}
		nodeInfo := &NodeInfo{
			NodeName:   node.Name,
			UniqueName: node.Name,
			Status:     statusOffline,
			IsManaged:  false,
			CreatedAt:  time.Now().Format(TimeFormat),
			UpdateAt:   time.Now().Format(TimeFormat),
		}
		if err := NodeServiceInstance().createNode(nodeInfo); err != nil {
			return err
		}
		hwlog.RunLog.Debugf("auto create unmanaged node %s", node.Name)
	}
	return nil
}
