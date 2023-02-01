// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"k8s.io/api/core/v1"

	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/types"
)

var (
	nodeNotFoundPattern = regexp.MustCompile(`nodes "([^"]+)" not found`)
)

// getNodeDetail get node detail
func getNodeDetail(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start get node detail")
	id, ok := input.(uint64)
	if !ok {
		hwlog.RunLog.Error("query node detail failed: para type not valid")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "query node detail convert param failed"}
	}
	var resp NodeInfoDetail
	nodeInfo, err := NodeServiceInstance().getNodeByID(id)
	if err != nil {
		hwlog.RunLog.Error("get node detail db query error")
		return common.RespMsg{Status: common.ErrorGetNode, Msg: "query node in db error", Data: nil}
	}
	resp.NodeInfo = *nodeInfo
	resp.NodeGroup, err = evalNodeGroup(id)
	if err != nil {
		hwlog.RunLog.Errorf("get node detail db query error, %s", err.Error())
		return common.RespMsg{Status: common.ErrorGetNode, Msg: err.Error(), Data: nil}
	}
	nodeResource, err := NodeStatusServiceInstance().GetAllocatableResource(nodeInfo.UniqueName)
	if err != nil {
		hwlog.RunLog.Warnf("get node detail query node resource error, %s", err.Error())
		nodeResource = &NodeResource{}
	}
	resp.NodeResource = *nodeResource
	resp.Status, err = NodeStatusServiceInstance().GetNodeStatus(nodeInfo.UniqueName)
	if err != nil {
		hwlog.RunLog.Warnf("get node detail query node status error, %s", err.Error())
		resp.Status = statusOffline
	}
	resp.Npu, err = NodeStatusServiceInstance().GetAllocatableNpu(nodeInfo.UniqueName)
	if err != nil {
		hwlog.RunLog.Warnf("get node detail query node npu error, %s", err.Error())
	}
	hwlog.RunLog.Info("node detail db query success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

func modifyNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start modify node")
	var req ModifyNodeReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("modify node convert request error, %s", err.Error())
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "", Data: nil}
	}
	if checkResult := newModifyNodeChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("modify node check parameters failed, %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}
	updatedColumns := map[string]interface{}{
		"NodeName":    req.NodeName,
		"Description": req.Description,
		"UpdatedAt":   time.Now().Format(TimeFormat),
	}
	if cnt, err := NodeServiceInstance().updateNode(*req.NodeID, managed, updatedColumns); err != nil || cnt != 1 {
		if err != nil && strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
			hwlog.RunLog.Error("node name is duplicate")
			return common.RespMsg{Status: common.ErrorNodeMrgDuplicate, Msg: "node name is duplicate", Data: nil}
		}
		hwlog.RunLog.Error("modify node db update error")
		return common.RespMsg{Status: common.ErrorModifyNode, Msg: "", Data: nil}
	}
	hwlog.RunLog.Info("modify node db update success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func getNodeStatistics(interface{}) common.RespMsg {
	hwlog.RunLog.Info("start get node statistics")
	nodes, err := NodeServiceInstance().listNodes()
	if err != nil {
		hwlog.RunLog.Error("failed to get node statistics, db query failed")
		return common.RespMsg{Status: common.ErrorCountNodeByStatus, Msg: ""}
	}
	statusMap := make(map[string]string)
	allNodeStatus := NodeStatusServiceInstance().ListNodeStatus()
	for hostname, status := range allNodeStatus {
		statusMap[hostname] = status
	}
	resp := make(map[string]int64)
	for _, node := range *nodes {
		status := statusOffline
		if nodeStatus, ok := statusMap[node.UniqueName]; ok {
			status = nodeStatus
		}
		if _, ok := resp[status]; !ok {
			resp[status] = 0
		}
		resp[status] += 1
	}
	hwlog.RunLog.Info("get node statistics success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

// ListNode get node list
func listManagedNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start list node managed")
	req, ok := input.(types.ListReq)
	if !ok {
		hwlog.RunLog.Error("list node convert request error")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "convert list request error", Data: nil}
	}
	total, err := NodeServiceInstance().countNodesByName(req.Name, managed)
	if err != nil {
		hwlog.RunLog.Error("count node failed")
		return common.RespMsg{Status: "", Msg: "count node failed", Data: nil}
	}
	nodes, err := NodeServiceInstance().listManagedNodesByName(req.PageNum, req.PageSize, req.Name)
	if err != nil {
		hwlog.RunLog.Error("list node failed")
		return common.RespMsg{Status: common.ErrorListNode, Msg: "list node failed", Data: nil}
	}
	var nodeList []NodeInfoExManaged
	for _, nodeInfo := range *nodes {
		var respItem NodeInfoExManaged
		respItem.NodeInfo = nodeInfo
		respItem.NodeGroup, err = evalNodeGroup(nodeInfo.ID)
		if err != nil {
			hwlog.RunLog.Errorf("list node db error: %s", err.Error())
			return common.RespMsg{Status: common.ErrorListNode, Msg: err.Error()}
		}
		respItem.Status, err = NodeStatusServiceInstance().GetNodeStatus(nodeInfo.UniqueName)
		if err != nil {
			respItem.Status = statusOffline
		}
		nodeList = append(nodeList, respItem)
	}
	resp := ListNodesResp{
		Nodes: nodeList,
		Total: int(total),
	}
	hwlog.RunLog.Info("list node success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

// ListNode get node list
func listUnmanagedNode(input interface{}) common.RespMsg {
	if err := autoAddUnmanagedNode(); err != nil {
		hwlog.RunLog.Error("auto add unmanaged node filed")
		return common.RespMsg{Status: common.ErrorListUnManagedNode, Msg: "auto add unmanaged node filed", Data: nil}
	}
	hwlog.RunLog.Info("start list node unmanaged")
	req, ok := input.(types.ListReq)
	if !ok {
		hwlog.RunLog.Error("list node convert request error")
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "", Data: nil}
	}
	total, err := NodeServiceInstance().countNodesByName(req.Name, unmanaged)
	if err != nil {
		hwlog.RunLog.Error("count node failed")
		return common.RespMsg{Status: "", Msg: "count node failed", Data: nil}
	}
	nodes, err := NodeServiceInstance().listUnManagedNodesByName(req.PageNum, req.PageSize, req.Name)
	if err == nil {
		var nodeList []NodeInfoEx
		for _, node := range *nodes {
			var respItem NodeInfoEx
			respItem.NodeInfo = node
			respItem.Status, err = NodeStatusServiceInstance().GetNodeStatus(node.UniqueName)
			if err != nil {
				respItem.Status = statusOffline
			}
			nodeList = append(nodeList, respItem)
		}
		resp := ListNodesUnmanagedResp{
			Nodes: nodeList,
			Total: int(total),
		}
		hwlog.RunLog.Info("list node unmanaged success")
		return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
	}
	if err == gorm.ErrRecordNotFound {
		hwlog.RunLog.Info("dont have any unmanaged node")
		return common.RespMsg{Status: common.Success, Msg: "dont have any unmanaged node", Data: nil}
	}
	hwlog.RunLog.Error("list unmanaged node failed")
	return common.RespMsg{Status: common.ErrorListUnManagedNode, Msg: "", Data: nil}
}

func autoAddUnmanagedNode() error {
	realNodes, err := kubeclient.GetKubeClient().ListNode()
	if err != nil {
		return err
	}
	dbNodeCount, err := GetTableCount(NodeInfo{})
	if err != nil {
		hwlog.RunLog.Error("get node table num failed")
		return err
	}
	// assume has one master node
	if len(realNodes.Items)-1 == dbNodeCount {
		return nil
	}
	for _, node := range realNodes.Items {
		if _, ok := node.Labels[masterNodeLabelKey]; ok {
			continue
		}
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
			IsManaged:  false,
			IP:         evalIpAddress(&node),
			CreatedAt:  time.Now().Format(TimeFormat),
			UpdatedAt:  time.Now().Format(TimeFormat),
		}
		if dbNodeCount >= maxNode {
			return errors.New("node number is enough, cannot create")
		}
		if err := NodeServiceInstance().createNode(nodeInfo); err != nil {
			return err
		}
		dbNodeCount += 1
		hwlog.RunLog.Debugf("auto create unmanaged node %s", node.Name)
	}
	return nil
}

func listNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start list all nodes")
	req, ok := input.(types.ListReq)
	if !ok {
		hwlog.RunLog.Error("list nodes convert request error")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "convert request error", Data: nil}
	}
	total, err := NodeServiceInstance().countAllNodesByName(req.Name)
	if err != nil {
		hwlog.RunLog.Error("count node failed")
		return common.RespMsg{Status: "", Msg: "count node failed", Data: nil}
	}
	nodes, err := NodeServiceInstance().listAllNodesByName(req.PageNum, req.PageSize, req.Name)
	if err != nil {
		hwlog.RunLog.Error("list all nodes failed")
		return common.RespMsg{Status: "", Msg: "list all nodes failed", Data: nil}
	}
	var nodeList []NodeInfoExManaged
	for _, nodeInfo := range *nodes {
		var respItem NodeInfoExManaged
		respItem.NodeGroup, err = evalNodeGroup(nodeInfo.ID)
		if err != nil {
			hwlog.RunLog.Errorf("list node id [%d] db error: %s", nodeInfo.ID, err.Error())
			return common.RespMsg{Msg: err.Error()}
		}
		respItem.Status, err = NodeStatusServiceInstance().GetNodeStatus(nodeInfo.UniqueName)
		if err != nil {
			respItem.Status = statusOffline
		}
		respItem.NodeInfo = nodeInfo
		nodeList = append(nodeList, respItem)
	}
	resp := ListNodesResp{
		Nodes: nodeList,
		Total: int(total),
	}
	hwlog.RunLog.Info("list all nodes success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

func batchDeleteNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start delete node")
	var req BatchDeleteNodeReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("failed to delete node, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error()}
	}
	if checkResult := newBatchDeleteNodeChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("failed to delete node, error: %v", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}
	var res types.BatchResp
	for _, nodeID := range *req.NodeIDs {
		if err := deleteSingleNode(nodeID); err != nil {
			hwlog.RunLog.Warnf("failed to delete node, error: err=%v", err)
			res.FailedIDs = append(res.FailedIDs, nodeID)
			continue
		}
		res.SuccessIDs = append(res.SuccessIDs, nodeID)
	}
	if len(res.FailedIDs) != 0 {
		return common.RespMsg{Status: common.ErrorDeleteNode, Data: res}
	}
	hwlog.RunLog.Info("delete node success")
	return common.RespMsg{Status: common.Success, Data: nil}
}

func deleteSingleNode(nodeID uint64) error {
	nodeInfo, err := NodeServiceInstance().getNodeByID(nodeID)
	if err != nil {
		return errors.New("db query failed")
	}
	if !nodeInfo.IsManaged {
		return errors.New("can't delete unmanaged node")
	}

	groupLabels := make([]string, 0, 4)
	node, err := kubeclient.GetKubeClient().GetNode(nodeInfo.UniqueName)
	if err != nil && isNodeNotFound(err) {
		hwlog.RunLog.Warnf("k8s query node failed, err=%v", err)
	} else if err != nil {
		return errors.New("k8s query node failed")
	} else {
		for _, label := range node.Labels {
			if strings.HasPrefix(label, common.NodeGroupLabelPrefix) {
				groupLabels = append(groupLabels, label)
			}
		}
	}
	err = kubeclient.GetKubeClient().DeleteNode(nodeInfo.UniqueName)
	if err != nil && isNodeNotFound(err) {
		hwlog.RunLog.Warnf("k8s delete node failed, err=%v", err)
	} else if err != nil {
		return errors.New("k8s delete node failed")
	}
	if _, err := NodeServiceInstance().deleteNodeByName(&NodeInfo{NodeName: nodeInfo.NodeName}); err != nil {
		return errors.New("db delete node failed")
	}
	if err = NodeServiceInstance().deleteRelationsToNode(nodeID); err != nil {
		return errors.New("db delete node relation failed")
	}
	return nil
}

func deleteNodeFromGroup(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start delete node from group")
	var req DeleteNodeToGroupReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("failed to delete from group, error: %v", err)
		return common.RespMsg{Msg: err.Error()}
	}
	if checkResult := newDeleteNodeFromGroupChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("failed to delete node from group, error: %v", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}
	var res types.BatchResp
	for _, nodeID := range *req.NodeIDs {
		if err := deleteSingleNodeRelation(*req.GroupID, nodeID); err != nil {
			hwlog.RunLog.Warnf("failed to delete node from group, error: err=%v", err)
			res.FailedIDs = append(res.FailedIDs, nodeID)
			continue
		}
		res.SuccessIDs = append(res.SuccessIDs, nodeID)
	}
	if len(res.FailedIDs) != 0 {
		return common.RespMsg{Status: common.ErrorDeleteNodeFromGroup, Msg: "", Data: res}
	}
	hwlog.RunLog.Info("delete node relation success")
	return common.RespMsg{Status: common.Success, Data: nil}
}

func batchDeleteNodeRelation(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start delete node relation")
	var req BatchDeleteNodeRelationReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("failed to delete node relation, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error()}
	}
	if checkResult := newBatchDeleteNodeRelationChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("failed to delete node relation, error: %v", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}
	var res types.BatchResp
	for _, relation := range req {
		if err := deleteSingleNodeRelation(*relation.GroupID, *relation.NodeID); err != nil {
			hwlog.RunLog.Warnf("failed to delete node relation, error: err=%v", err)
			res.FailedIDs = append(res.FailedIDs, relation)
			continue
		}
		res.SuccessIDs = append(res.SuccessIDs, relation)
	}
	if len(res.FailedIDs) != 0 {
		return common.RespMsg{Status: common.ErrorDeleteNodeFromGroup, Msg: "", Data: res}
	}
	hwlog.RunLog.Info("delete node relation success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func deleteSingleNodeRelation(groupID, nodeID uint64) error {
	nodeInfo, err := NodeServiceInstance().getNodeByID(nodeID)
	if err != nil {
		return errors.New("db query failed")
	}
	rowsAffected, err := NodeServiceInstance().deleteNodeToGroup(&NodeRelation{NodeID: nodeID, GroupID: groupID})
	if err != nil {
		return errors.New("db delete failed")
	}
	if rowsAffected < 1 {
		return errors.New("no such relation")
	}
	nodeLabel := fmt.Sprintf("%s%d", common.NodeGroupLabelPrefix, groupID)
	_, err = kubeclient.GetKubeClient().DeleteNodeLabels(nodeInfo.UniqueName, []string{nodeLabel})
	if err != nil && isNodeNotFound(err) {
		hwlog.RunLog.Warnf("k8s delete label failed, err=%v", err)
	} else if err != nil {
		return errors.New("k8s delete label failed")
	}
	return nil
}

func isNodeNotFound(err error) bool {
	return nodeNotFoundPattern.MatchString(err.Error())
}

func evalNodeGroup(nodeID uint64) (string, error) {
	relations, err := NodeServiceInstance().getRelationsByNodeID(nodeID)
	if err != nil {
		hwlog.RunLog.Error("get node detail db query error")
		return "", errors.New("db query error")
	}
	nodeGroupName := ""
	if len(*relations) > 0 {
		var buffer bytes.Buffer
		for index, relation := range *relations {
			nodeGroup, err := NodeServiceInstance().getNodeGroupByID(relation.GroupID)
			if err != nil {
				hwlog.RunLog.Error("get node detail db query error")
				return "", errors.New("db query error")
			}
			if index != 0 {
				buffer.WriteString(",")
			}
			buffer.WriteString(nodeGroup.GroupName)
		}
		nodeGroupName = buffer.String()
	}
	return nodeGroupName, nil
}

func addNodeRelation(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start add node to group")
	var req AddNodeToGroupReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("add node to group convert request error, %s", err.Error())
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "", Data: nil}
	}
	if checkResult := newAddNodeRelationChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("failed to add node to group, error: %v", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}

	checker := specificationChecker{nodeService: NodeServiceInstance()}
	if err := checker.checkAddNodeToGroup(*req.NodeIDs, []uint64{*req.GroupID}); err != nil {
		hwlog.RunLog.Errorf("add node to group check spec error: %s", err.Error())
		return common.RespMsg{Status: common.ErrorCheckNodeMrgSize, Msg: err.Error()}
	}
	res, err := addNode(req)
	if err != nil {
		hwlog.RunLog.Error(err.Error())
		return common.RespMsg{Status: common.ErrorAddNodeToGroup, Msg: err.Error(), Data: res}
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func addNode(req AddNodeToGroupReq) (*types.BatchResp, error) {
	var res types.BatchResp
	nodeGroup, err := NodeServiceInstance().getNodeGroupByID(*req.GroupID)
	if err != nil {
		return nil, fmt.Errorf("dont have this node group id(%d)", *req.GroupID)
	}
	for i, id := range *req.NodeIDs {
		nodeDb, err := NodeServiceInstance().getManagedNodeByID(id)
		if err != nil {
			res.FailedIDs = append(res.FailedIDs, id)
			hwlog.RunLog.Errorf("no found node id %d", id)
			continue
		}
		relation := NodeRelation{
			NodeID:    (*req.NodeIDs)[i],
			GroupID:   *req.GroupID,
			CreatedAt: time.Now().Format(TimeFormat)}
		if err := nodeServiceInstance.addNodeToGroup(&[]NodeRelation{relation}); err != nil {
			res.FailedIDs = append(res.FailedIDs, id)
			hwlog.RunLog.Errorf("add node(%s) to group(%d) to db error", nodeDb.NodeName, nodeGroup.ID)
			continue
		}
		label := map[string]string{fmt.Sprintf("%s%d", common.NodeGroupLabelPrefix, nodeGroup.ID): ""}
		if _, err := kubeclient.GetKubeClient().AddNodeLabels(nodeDb.UniqueName, label); err != nil {
			res.FailedIDs = append(res.FailedIDs, id)
			hwlog.RunLog.Errorf("node(%s) patch label(%d) error", nodeDb.NodeName, nodeGroup.ID)
			continue
		}
		res.SuccessIDs = append(res.SuccessIDs, id)
	}
	if len(res.FailedIDs) != 0 {
		return &res, errors.New("add some nodes to group failed")
	}
	return &res, nil
}

func addUnManagedNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start add unmanaged node")
	var req AddUnManagedNodeReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("add unmanaged node convert request error, %s", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "convert request error", Data: nil}
	}
	if checkResult := newAddUnManagedNodeChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("add unmanaged node validate parameters error, %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}

	checker := specificationChecker{nodeService: NodeServiceInstance()}
	if err := checker.checkAddNodeToGroup([]uint64{*req.NodeID}, req.GroupIDs); err != nil {
		hwlog.RunLog.Errorf("add unmanaged node to group check spec error: %s", err.Error())
		return common.RespMsg{Status: common.ErrorCheckNodeMrgSize, Msg: err.Error()}
	}
	updatedColumns := map[string]interface{}{
		"NodeName":    req.NodeName,
		"Description": req.Description,
		"IsManaged":   managed,
		"UpdatedAt":   time.Now().Format(TimeFormat),
	}
	if cnt, err := NodeServiceInstance().updateNode(*req.NodeID, unmanaged, updatedColumns); err != nil || cnt != 1 {
		hwlog.RunLog.Error("add unmanaged node error")
		return common.RespMsg{Status: common.ErrorAddUnManagedNode, Msg: "add node to mef system error", Data: nil}
	}
	var addNodeRes types.BatchResp
	for _, id := range req.GroupIDs {
		addReq := AddNodeToGroupReq{NodeIDs: &[]uint64{*req.NodeID}, GroupID: &id}
		_, err := addNode(addReq)
		if err != nil {
			addNodeRes.FailedIDs = append(addNodeRes.FailedIDs, id)
			hwlog.RunLog.Error(err)
			continue
		}
		addNodeRes.SuccessIDs = append(addNodeRes.SuccessIDs, id)
	}
	if len(addNodeRes.FailedIDs) != 0 {
		return common.RespMsg{Status: common.ErrorAddUnManagedNode,
			Msg: "add node to mef success, but node cannot join some group", Data: addNodeRes}
	}
	hwlog.RunLog.Info("add unmanaged node success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func evalIpAddress(node *v1.Node) string {
	var ipAddresses []string
	for _, addr := range node.Status.Addresses {
		if addr.Type != v1.NodeExternalIP && addr.Type != v1.NodeInternalIP {
			continue
		}
		ipAddresses = append(ipAddresses, addr.Address)
	}
	return strings.Join(ipAddresses, ",")
}

func innerGetNodeInfoByUniqueName(input interface{}) common.RespMsg {
	req, ok := input.(types.InnerGetNodeInfoByNameReq)
	if !ok {
		hwlog.RunLog.Error("parse inner message content failed")
		return common.RespMsg{Status: "", Msg: "parse inner message content failed", Data: nil}
	}
	nodeInfo, err := NodeServiceInstance().getNodeByUniqueName(req.UniqueName)
	if err != nil {
		hwlog.RunLog.Error("get node info by unique name failed")
		return common.RespMsg{Status: "", Msg: "get node info by unique name failed", Data: nil}
	}
	resp := types.InnerGetNodeInfoByNameResp{
		NodeID:   nodeInfo.ID,
		NodeName: nodeInfo.NodeName,
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

func innerGetNodeStatus(input interface{}) common.RespMsg {
	req, ok := input.(types.InnerGetNodeStatusReq)
	if !ok {
		hwlog.RunLog.Error("parse inner message content failed")
		return common.RespMsg{Status: "", Msg: "parse inner message content failed", Data: nil}
	}
	status, err := NodeStatusServiceInstance().GetNodeStatus(req.UniqueName)
	if err != nil {
		hwlog.RunLog.Error("inner message get node status failed")
		return common.RespMsg{Status: "", Msg: "internal get node status failed", Data: nil}
	}
	resp := types.InnerGetNodeStatusResp{
		NodeStatus: status,
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}
