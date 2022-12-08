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

	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/util"
	"huawei.com/mindxedge/base/common"
)

var nodeNotFoundPattern = regexp.MustCompile(`nodes "([^"]+)" not found`)

// CreateNode Create Node
func createNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start create node")
	var req CreateEdgeNodeReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("create node convert request error, %s", err.Error())
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	if err := req.Check(); err != nil {
		hwlog.RunLog.Errorf("create node validate parameters error: %s", err.Error())
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: err.Error()}
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
	var req GetNodeDetailReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("get node detail convert request error, %s", err.Error())
		return common.RespMsg{Status: "", Msg: "convert request error", Data: nil}
	}
	if err := req.Check(); err != nil {
		hwlog.RunLog.Errorf("get node detail check parameters failed, %s", err.Error())
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: err.Error()}
	}
	nodeInfo, err := NodeServiceInstance().GetNodeByID(req.Id)
	if err != nil {
		hwlog.RunLog.Error("get node detail db query error")
		return common.RespMsg{Status: "", Msg: "db query error", Data: nil}
	}
	nodeGroupName, err := joinNodeGroups(req.Id)
	if err != nil {
		hwlog.RunLog.Errorf("get node detail db query error, %s", err.Error())
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	status := NodeStatusServiceInstance().GetNodeStatus(nodeInfo.UniqueName)
	resp := GetNodeDetailResp{
		Id:          nodeInfo.ID,
		NodeName:    nodeInfo.NodeName,
		UniqueName:  nodeInfo.UniqueName,
		Description: nodeInfo.Description,
		Status:      status,
		CreatedAt:   nodeInfo.CreatedAt,
		UpdatedAt:   nodeInfo.UpdateAt,
		Cpu:         nodeInfo.CPUCore,
		Memory:      nodeInfo.Memory,
		Npu:         nodeInfo.NPUType,
		NodeType:    nodeInfo.NodeType,
		NodeGroup:   nodeGroupName,
	}
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
	if err := req.Check(); err != nil {
		hwlog.RunLog.Errorf("modify node check parameters failed, %s", err.Error())
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: err.Error()}
	}
	updatedColumns := map[string]interface{}{
		"NodeName":    req.NodeName,
		"Description": req.Description,
		"UpdateAt":    time.Now().Format(TimeFormat),
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
	resp := map[string]int64{
		statusReady:    0,
		statusNotReady: 0,
		statusUnknown:  0,
		statusOffline:  0,
	}
	nodes, err := NodeServiceInstance().listNodes()
	if err != nil {
		hwlog.RunLog.Error("failed to get node statistics, db query failed")
		return common.RespMsg{Msg: "db query failed"}
	}
	statusMap := NodeStatusServiceInstance().ListNodeStatus()
	for _, node := range *nodes {
		status := statusOffline
		if nodeStatus, ok := statusMap[node.UniqueName]; ok {
			status = nodeStatus
		}
		if _, ok := resp[status]; !ok {
			continue
		}
		resp[status] += 1
	}
	hwlog.RunLog.Info("get node statistics success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

// ListNode get node list
func listNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start list node managed")
	req, ok := input.(util.ListReq)
	if !ok {
		hwlog.RunLog.Info("list node convert request error")
		return common.RespMsg{Status: "", Msg: "convert request error", Data: nil}
	}
	nodes, err := NodeServiceInstance().listNodesByName(req.PageNum, req.PageSize, req.Name)
	if err == nil {
		for i := range *nodes {
			nodePtr := &(*nodes)[i]
			nodePtr.Status = NodeStatusServiceInstance().GetNodeStatus(nodePtr.UniqueName)
		}
		resp := ListNodesResp{
			Nodes: nodes,
			Total: len(*nodes),
		}
		hwlog.RunLog.Info("list node success")
		return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
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
	req, ok := input.(util.ListReq)
	if !ok {
		hwlog.RunLog.Info("list node convert request error")
		return common.RespMsg{Status: "", Msg: "convert request error", Data: nil}
	}

	nodes, err := NodeServiceInstance().listUnManagedNodesByName(req.PageNum, req.PageSize, req.Name)
	if err == nil {
		for i := range *nodes {
			nodePtr := &(*nodes)[i]
			nodePtr.Status = NodeStatusServiceInstance().GetNodeStatus(nodePtr.UniqueName)
		}
		resp := ListNodesResp{
			Nodes: nodes,
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
	nodeDb, err := GetTableCount(NodeInfo{})
	if err != nil {
		hwlog.RunLog.Error("get node table num failed")
		return err
	}
	if len(realNodes.Items) == nodeDb {
		return nil
	}
	for _, node := range realNodes.Items {
		_, err := NodeServiceInstance().GetNodeByUniqueName(node.Name)
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

func batchDeleteNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start delete node")
	var req BatchDeleteNodeReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("failed to delete node, error: %v", err)
		return common.RespMsg{Msg: err.Error()}
	}
	if err := req.Check(); err != nil {
		hwlog.RunLog.Errorf("failed to delete node, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: err.Error()}
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
	nodeInfo, err := NodeServiceInstance().GetNodeByID(nodeID)
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
	if err = NodeServiceInstance().deleteNodeByName(&NodeInfo{NodeName: nodeInfo.NodeName}); err != nil {
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
	if err := req.Check(); err != nil {
		hwlog.RunLog.Errorf("failed to delete node relation, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: err.Error()}
	}
	var deleteCount int64
	for _, relation := range req {
		if err := deleteSingleNodeRelation(relation.GroupID, relation.NodeID); err != nil {
			hwlog.RunLog.Warnf("failed to delete node relation, error: err=%v", err)
			continue
		}
		deleteCount += 1
	}
	hwlog.RunLog.Info("delete node relation success")
	return common.RespMsg{Status: common.Success, Data: deleteCount}
}

func deleteSingleNodeRelation(groupID, nodeID int64) error {
	nodeInfo, err := NodeServiceInstance().GetNodeByID(nodeID)
	if err != nil {
		return errors.New("db query failed")
	}
	rowsAffected, err := NodeServiceInstance().deleteRelation(&NodeRelation{NodeID: nodeID, GroupID: groupID})
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

func joinNodeGroups(nodeID int64) (string, error) {
	relations, err := NodeServiceInstance().getRelationsByNodeID(nodeID)
	if err != nil {
		hwlog.RunLog.Error("get node detail db query error")
		return "", errors.New("db query error")
	}
	nodeGroupName := ""
	if len(*relations) > 0 {
		var buffer bytes.Buffer
		for index, relation := range *relations {
			nodeGroup, err := NodeServiceInstance().GetNodeGroupByID(relation.GroupID)
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
	total, err := NodeServiceInstance().countNodeByGroup(req.GroupID)
	if err != nil {
		hwlog.RunLog.Error("get node in group table num failed")
		return common.RespMsg{Status: "", Msg: "get node in group table num failed", Data: nil}
	}
	num := total + int64(len(req.NodeID))
	if num > maxNodePerGroup {
		hwlog.RunLog.Error("node in group number is enough, connot create")
		return common.RespMsg{Status: "", Msg: "node in group number is enough, connot create", Data: nil}
	}

	if err := addNode(req); err != nil {
		hwlog.RunLog.Warn(err.Error())
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func addNode(req AddNodeToGroupReq) error {
	var errorNode string
	nodeGroup, err := NodeServiceInstance().GetNodeGroupByID(req.GroupID)
	if err != nil {
		return errors.New("dont have this node group")
	}
	var nodeRelation []NodeRelation
	for i, id := range req.NodeID {
		nodeDb, err := NodeServiceInstance().getManagedNodeByID(id)
		if err != nil {
			errorNode = fmt.Sprintf("%d,%s", id, errorNode)
			hwlog.RunLog.Errorf("no found node id %d", id)
			continue
		}
		label := make(map[string]string)
		label[fmt.Sprintf("%s%d", common.NodeGroupLabelPrefix, nodeGroup.ID)] = ""
		if _, err := kubeclient.GetKubeClient().AddNodeLabels(nodeDb.UniqueName, label); err != nil {
			hwlog.RunLog.Errorf("node(%s) patch label(%d) error", nodeDb.NodeName, nodeGroup.ID)
			continue
		}
		relation := NodeRelation{
			NodeID:    req.NodeID[i],
			GroupID:   req.GroupID,
			CreatedAt: time.Now().Format(TimeFormat)}
		nodeRelation = append(nodeRelation, relation)
	}
	if len(nodeRelation) == 0 {
		return errors.New("all node failed to join group")
	}
	if err = nodeServiceInstance.addNodeToGroup(&nodeRelation); err != nil {
		return errors.New("add node relations to db error")
	}
	if errorNode != "" {
		return fmt.Errorf("not fount node id:%s", errorNode)
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
	if err := req.Check(); err != nil {
		hwlog.RunLog.Errorf("add unmanaged node validate parameters error, %s", err.Error())
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: err.Error()}
	}
	// todo nodePerGroup and groupPerNode  limit count
	updatedColumns := map[string]interface{}{
		"NodeName":    req.NodeName,
		"Description": req.Description,
		"IsManaged":   managed,
		"UpdateAt":    time.Now().Format(TimeFormat),
	}
	if err := NodeServiceInstance().updateNode(req.NodeID, updatedColumns); err != nil {
		hwlog.RunLog.Error("add unmanaged node error")
		return common.RespMsg{Status: "", Msg: "add unmanaged node error", Data: nil}
	}
	var errorGroup string
	for _, id := range req.GroupID {
		addReq := AddNodeToGroupReq{NodeID: []int64{req.NodeID}, GroupID: id}
		if err := addNode(addReq); err != nil {
			errorGroup = fmt.Sprintf("%d,%s", id, errorGroup)
		}
	}
	if errorGroup != "" {
		return common.RespMsg{Status: "", Msg: fmt.Sprintf("cannot join group:%s", errorGroup), Data: nil}
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func batchDeleteNodeGroup(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start batch delete node group")
	var req BatchDeleteNodeGroupReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("batch delete node group convert request error, %s", err)
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	var delSuccessGroupID []int64
	for _, groupID := range req.GroupID {
		nodeGroup, err := NodeServiceInstance().GetNodeGroupByID(groupID)
		if err != nil {
			hwlog.RunLog.Errorf("get node group by group id %d failed", groupID)
			continue
		}
		relations, err := NodeServiceInstance().listNodeRelationsByGroupId(groupID)
		if err != nil {
			hwlog.RunLog.Errorf("get relations between node and node group by group id %d failed", groupID)
			continue
		}
		var operationSuccessTimes int64
		for _, relation := range *relations {
			if err := deleteSingleNodeRelation(nodeGroup.ID, relation.NodeID); err != nil {
				hwlog.RunLog.Errorf("patch node state failed:%v", err)
				continue
			}
			if err = NodeServiceInstance().deleteNodeToGroup(&relation); err != nil {
				hwlog.RunLog.Errorf("delete relation %v from db failed:%v", relation, err)
				continue
			}
			operationSuccessTimes++
		}
		if operationSuccessTimes == int64(len(*relations)) {
			if err = NodeServiceInstance().deleteNodeGroup(groupID); err != nil {
				hwlog.RunLog.Errorf("delete node group by group id %d failed:%v", groupID, err)
				continue
			}
			delSuccessGroupID = append(delSuccessGroupID, groupID)
		}
	}
	if len(delSuccessGroupID) > 0 {
		return common.RespMsg{Status: common.Success, Msg: "batch delete node group success", Data: delSuccessGroupID}
	}
	return common.RespMsg{Status: "", Msg: "batch delete node group failed", Data: nil}
}
