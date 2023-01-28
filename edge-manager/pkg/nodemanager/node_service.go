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
	"edge-manager/pkg/util"
)

var (
	nodeNotFoundPattern = regexp.MustCompile(`nodes "([^"]+)" not found`)
)

// CreateNode Create Node
func createNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start create node")
	var req CreateEdgeNodeReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("create node convert request error, %s", err.Error())
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	if checkResult := newCreateEdgeNodeChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("create node validate parameters error: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}

	checker := specificationChecker{nodeService: NodeServiceInstance()}
	if err := checker.checkAddNodes(1); err != nil {
		hwlog.RunLog.Errorf("create node check spec error: %s", err.Error())
		return common.RespMsg{Msg: err.Error()}
	}
	node := &NodeInfo{
		Description: req.Description,
		UniqueName:  *req.UniqueName,
		NodeName:    *req.NodeName,
		IsManaged:   true,
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdatedAt:   time.Now().Format(TimeFormat),
	}
	if err := NodeServiceInstance().createNode(node); err != nil {
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
	id, ok := input.(int64)
	if !ok {
		hwlog.RunLog.Error("query node detail failed: para type not valid")
		return common.RespMsg{Status: "", Msg: "query node detail failed", Data: nil}
	}
	nodeInfo, err := NodeServiceInstance().getNodeByID(id)
	if err != nil {
		hwlog.RunLog.Error("get node detail db query error")
		return common.RespMsg{Status: "", Msg: "db query error", Data: nil}
	}
	nodeGroupName, err := evalNodeGroup(id)
	if err != nil {
		hwlog.RunLog.Errorf("get node detail db query error, %s", err.Error())
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	nodeInfoDynamic, _ := NodeStatusServiceInstance().Get(nodeInfo.UniqueName)
	var resp NodeInfoDetail
	resp.Extend(nodeInfo, nodeInfoDynamic, nodeGroupName)
	hwlog.RunLog.Info("node detail db query success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

func modifyNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start modify node")
	var req ModifyNodeReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("modify node convert request error, %s", err.Error())
		return common.RespMsg{Status: "", Msg: "convert request error", Data: nil}
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
	nodes, err := NodeServiceInstance().listNodes()
	if err != nil {
		hwlog.RunLog.Error("failed to get node statistics, db query failed")
		return common.RespMsg{Msg: "db query failed"}
	}
	nodeInfoDynamics := NodeStatusServiceInstance().List()
	statusMap := make(map[string]string)
	for _, dynamic := range *nodeInfoDynamics {
		statusMap[dynamic.Hostname] = dynamic.Status
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

func getGroupNodeStatistics(interface{}) common.RespMsg {
	hwlog.RunLog.Info("start get node group statistics")
	total, err := GetTableCount(NodeGroup{})
	if err != nil {
		hwlog.RunLog.Error("failed to get node group statistics, db query failed")
		return common.RespMsg{Msg: "db query failed"}
	}
	resp := map[string]interface{}{
		"total": total,
	}
	hwlog.RunLog.Info("get node group statistics success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

// ListNode get node list
func listManagedNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start list node managed")
	req, ok := input.(util.ListReq)
	if !ok {
		hwlog.RunLog.Error("list node convert request error")
		return common.RespMsg{Status: "", Msg: "convert request error", Data: nil}
	}
	nodes, err := NodeServiceInstance().listManagedNodesByName(req.PageNum, req.PageSize, req.Name)
	if err != nil {
		hwlog.RunLog.Error("list node failed")
		return common.RespMsg{Status: "", Msg: "list node failed", Data: nil}
	}
	var nodeInfoDetails []NodeInfoDetail
	for _, nodeInfo := range *nodes {
		nodeGroup, err := evalNodeGroup(nodeInfo.ID)
		if err != nil {
			hwlog.RunLog.Errorf("list node db error: %s", err.Error())
			return common.RespMsg{Msg: err.Error()}
		}
		nodeInfoDynamic, _ := NodeStatusServiceInstance().Get(nodeInfo.UniqueName)
		var nodeInfoDetail NodeInfoDetail
		nodeInfoDetail.Extend(&nodeInfo, nodeInfoDynamic, nodeGroup)
		nodeInfoDetails = append(nodeInfoDetails, nodeInfoDetail)
	}
	resp := ListNodesResp{
		Nodes: &nodeInfoDetails,
		Total: len(*nodes),
	}
	hwlog.RunLog.Info("list node success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

// ListNode get node list
func listUnmanagedNode(input interface{}) common.RespMsg {
	if err := autoAddUnmanagedNode(); err != nil {
		hwlog.RunLog.Error("auto add unmanaged node filed")
		return common.RespMsg{Status: "", Msg: "auto add unmanaged node filed", Data: nil}
	}
	hwlog.RunLog.Info("start list node unmanaged")
	req, ok := input.(util.ListReq)
	if !ok {
		hwlog.RunLog.Error("list node convert request error")
		return common.RespMsg{Status: "", Msg: "convert request error", Data: nil}
	}
	nodes, err := NodeServiceInstance().listUnManagedNodesByName(req.PageNum, req.PageSize, req.Name)
	if err == nil {
		var nodeInfoExs []NodeInfoEx
		for _, node := range *nodes {
			nodeInfoDynamic, _ := NodeStatusServiceInstance().Get(node.UniqueName)
			var nodeInfoEx NodeInfoEx
			nodeInfoEx.Extend(&node, nodeInfoDynamic)
			nodeInfoExs = append(nodeInfoExs, nodeInfoEx)
		}
		resp := ListNodesUnmanagedResp{
			Nodes: &nodeInfoExs,
			Total: len(*nodes),
		}
		hwlog.RunLog.Info("list node unmanaged success")
		return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
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

func batchDeleteNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start delete node")
	var req BatchDeleteNodeReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("failed to delete node, error: %v", err)
		return common.RespMsg{Msg: err.Error()}
	}
	if checkResult := newBatchDeleteNodeChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("failed to delete node, error: %v", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}
	var deleteCount int64
	for _, nodeID := range req {
		if err := deleteSingleNode(nodeID); err != nil {
			hwlog.RunLog.Warnf("failed to delete node, error: err=%v", err)
			continue
		}
		deleteCount += 1
	}
	hwlog.RunLog.Info("delete node success")
	return common.RespMsg{Status: common.Success, Data: deleteCount}
}

func deleteSingleNode(nodeID int64) error {
	nodeInfo, err := NodeServiceInstance().getNodeByID(nodeID)
	if err != nil {
		return errors.New("db query failed")
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
	if _, err := NodeServiceInstance().deleteNodeByName(&NodeInfo{NodeName: nodeInfo.NodeName}); err != nil {
		return errors.New("db delete failed")
	}
	if err = NodeServiceInstance().deleteRelationsToNode(nodeID); err != nil {
		return errors.New("db delete failed")
	}
	if len(groupLabels) > 0 {
		_, err = kubeclient.GetKubeClient().DeleteNodeLabels(nodeInfo.UniqueName, groupLabels)
		if err != nil && isNodeNotFound(err) {
			hwlog.RunLog.Warnf("k8s delete label failed, err=%v", err)
		} else if err != nil {
			return errors.New("k8s delete label failed")
		}
	}
	err = kubeclient.GetKubeClient().DeleteNode(nodeInfo.UniqueName)
	if err != nil && isNodeNotFound(err) {
		hwlog.RunLog.Warnf("k8s delete node failed, err=%v", err)
	} else if err != nil {
		return errors.New("k8s delete node failed")
	}
	return nil
}

func batchDeleteNodeRelation(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start delete node relation")
	var req BatchDeleteNodeRelationReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("failed to delete node relation, error: %v", err)
		return common.RespMsg{Msg: err.Error()}
	}
	if checkResult := newBatchDeleteNodeRelationChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("failed to delete node relation, error: %v", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}
	var deleteCount int64
	for _, relation := range req {
		if err := deleteSingleNodeRelation(*relation.GroupID, *relation.NodeID); err != nil {
			hwlog.RunLog.Warnf("failed to delete node relation, error: err=%v", err)
			continue
		}
		deleteCount += 1
	}
	hwlog.RunLog.Info("delete node relation success")
	return common.RespMsg{Status: common.Success, Data: deleteCount}
}

func deleteSingleNodeRelation(groupID, nodeID int64) error {
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

func evalNodeGroup(nodeID int64) (string, error) {
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
		return common.RespMsg{Status: "", Msg: "convert request error", Data: nil}
	}
	if checkResult := newAddNodeRelationChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("failed to add node to group, error: %v", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}

	checker := specificationChecker{nodeService: NodeServiceInstance()}
	if err := checker.checkAddNodeToGroup(*req.NodeIDs, []int64{*req.GroupID}); err != nil {
		hwlog.RunLog.Errorf("add node to group check spec error: %s", err.Error())
		return common.RespMsg{Msg: err.Error()}
	}
	if err := addNode(req); err != nil {
		hwlog.RunLog.Warn(err.Error())
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func addNode(req AddNodeToGroupReq) error {
	var errorNode string
	nodeGroup, err := NodeServiceInstance().getNodeGroupByID(*req.GroupID)
	if err != nil {
		return errors.New("dont have this node group")
	}
	var nodeRelation []NodeRelation
	for i, id := range *req.NodeIDs {
		nodeDb, err := NodeServiceInstance().getManagedNodeByID(id)
		if err != nil {
			errorNode = fmt.Sprintf("%d,%s", id, errorNode)
			hwlog.RunLog.Errorf("no found node id %d", id)
			continue
		}
		relation := NodeRelation{
			NodeID:    (*req.NodeIDs)[i],
			GroupID:   *req.GroupID,
			CreatedAt: time.Now().Format(TimeFormat)}
		if err := nodeServiceInstance.addNodeToGroup(&[]NodeRelation{relation}); err != nil {
			errorNode = fmt.Sprintf("%d,%s", id, errorNode)
			hwlog.RunLog.Errorf("add node(%s) to group(%d) to db error", nodeDb.NodeName, nodeGroup.ID)
			continue
		}
		label := map[string]string{fmt.Sprintf("%s%d", common.NodeGroupLabelPrefix, nodeGroup.ID): ""}
		if _, err := kubeclient.GetKubeClient().AddNodeLabels(nodeDb.UniqueName, label); err != nil {
			errorNode = fmt.Sprintf("%d,%s", id, errorNode)
			hwlog.RunLog.Errorf("node(%s) patch label(%d) error", nodeDb.NodeName, nodeGroup.ID)
			continue
		}
		nodeRelation = append(nodeRelation, relation)
	}
	if len(nodeRelation) == 0 {
		return errors.New("all node failed to join group")
	}
	if errorNode != "" {
		return fmt.Errorf("some node failed to join group:%s", errorNode)
	}
	return nil
}

func addUnManagedNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start add unmanaged node")
	var req AddUnManagedNodeReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("add unmanaged node convert request error, %s", err)
		return common.RespMsg{Status: "", Msg: "convert request error", Data: nil}
	}
	if checkResult := newAddUnManagedNodeChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("add unmanaged node validate parameters error, %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}

	checker := specificationChecker{nodeService: NodeServiceInstance()}
	if err := checker.checkAddNodeToGroup([]int64{*req.NodeID}, req.GroupIDs); err != nil {
		hwlog.RunLog.Errorf("add unmanaged node to group check spec error: %s", err.Error())
		return common.RespMsg{Msg: err.Error()}
	}
	updatedColumns := map[string]interface{}{
		"NodeName":    req.NodeName,
		"Description": req.Description,
		"IsManaged":   managed,
		"UpdatedAt":   time.Now().Format(TimeFormat),
	}
	if cnt, err := NodeServiceInstance().updateNode(*req.NodeID, unmanaged, updatedColumns); err != nil || cnt != 1 {
		hwlog.RunLog.Error("add unmanaged node error")
		return common.RespMsg{Status: "", Msg: "add unmanaged node error", Data: nil}
	}
	var addNodeError string
	for _, id := range req.GroupIDs {
		addReq := AddNodeToGroupReq{NodeIDs: &[]int64{*req.NodeID}, GroupID: &id}
		if err := addNode(addReq); err != nil {
			errorMessage := fmt.Sprintf("cannot join group:group=%d, node=%d", id, req.NodeID)
			addNodeError = errorMessage
			hwlog.RunLog.Warn(errorMessage)
		}
	}
	if addNodeError != "" {
		return common.RespMsg{Status: "", Msg: addNodeError, Data: nil}
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
	nodeInfoDynamic, ok := NodeStatusServiceInstance().Get(req.UniqueName)
	if !ok {
		hwlog.RunLog.Error("inner message get node status failed")
		return common.RespMsg{Status: "", Msg: "internal get node status failed", Data: nil}
	}
	resp := types.InnerGetNodeStatusResp{
		NodeStatus: nodeInfoDynamic.Status,
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}
