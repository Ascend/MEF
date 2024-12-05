// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"edge-manager/pkg/config"
	"edge-manager/pkg/constants"
	"edge-manager/pkg/types"
	"edge-manager/pkg/util"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/logmgmt"
)

const maxNodeInfos = 2048

var (
	nodeNotFoundPattern = regexp.MustCompile(`nodes "([^"]+)" not found`)
)

func getNodeDetail(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start get node detail")
	var req map[string]interface{}
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("query node detail failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed"}
	}

	var handleFunc = map[string]func(interface{}) common.RespMsg{
		constants.IdKey: getNodeDetailById,
		constants.SnKey: getNodeDetailBySn,
	}

	identifier, ok := req[constants.KeySymbol]
	if !ok {
		hwlog.RunLog.Error("received unsupported key")
		return common.RespMsg{Status: common.ErrorGetNode, Msg: "unsupported msg key received", Data: nil}
	}

	strIdentifier, ok := identifier.(string)
	if !ok {
		hwlog.RunLog.Error("received identifier type is not string")
		return common.RespMsg{Status: common.ErrorGetNode, Msg: "received identifier type is not string", Data: nil}
	}

	dealer, ok := handleFunc[strIdentifier]
	if !ok {
		hwlog.RunLog.Error("received unsupported identifier")
		return common.RespMsg{Status: common.ErrorGetNode, Msg: "unsupported identifier received", Data: nil}
	}
	value, ok := req[constants.ValueSymbol]
	if !ok {
		hwlog.RunLog.Errorf("received unsupported value")
		return common.RespMsg{Status: common.ErrorGetNode, Msg: "unsupported msg value received", Data: nil}
	}

	return dealer(value)
}

func getNodeDetailById(input interface{}) common.RespMsg {
	idFloat, ok := input.(float64)
	if !ok {
		hwlog.RunLog.Error("query node detail failed: id para type not valid")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "query node detail convert param failed"}
	}
	id := uint64(idFloat)

	if checkResult := newGetNodeDetailIdChecker().Check(id); !checkResult.Result {
		hwlog.RunLog.Errorf("query node detail parameters failed, %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid,
			Msg: fmt.Sprintf("id check failed: %s", checkResult.Reason)}
	}
	var resp NodeInfoDetail
	nodeInfo, err := NodeServiceInstance().getNodeByID(id)
	if err != nil {
		hwlog.RunLog.Error("get node detail by id db query error")
		return common.RespMsg{Status: common.ErrorGetNode, Msg: "query node in db error", Data: nil}
	}
	resp.NodeInfo = *nodeInfo
	extResp, err := setNodeExtInfos(resp)
	if err != nil {
		return common.RespMsg{Status: common.ErrorGetNode, Msg: err.Error(), Data: nil}
	}

	hwlog.RunLog.Info("node detail db query success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: extResp}
}

func getNodeDetailBySn(input interface{}) common.RespMsg {
	sn, ok := input.(string)
	if !ok {
		hwlog.RunLog.Error("query node detail failed: sn para type not valid")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "query node detail convert param failed"}
	}

	snChecker := checker.GetSnChecker("", true)
	if checkResult := snChecker.Check(sn); !checkResult.Result {
		hwlog.RunLog.Errorf("query node detail parameters failed, %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid,
			Msg: fmt.Sprintf("sn check failed: %s", checkResult.Reason)}
	}

	var resp NodeInfoDetail
	nodeInfo, err := NodeServiceInstance().getNodeBySn(sn)
	if err != nil {
		hwlog.RunLog.Error("get node detail by sn db query error")
		return common.RespMsg{Status: common.ErrorGetNode, Msg: "query node in db error", Data: nil}
	}
	resp.NodeInfo = *nodeInfo
	extResp, err := setNodeExtInfos(resp)
	if err != nil {
		return common.RespMsg{Status: common.ErrorGetNode, Msg: err.Error(), Data: nil}
	}

	hwlog.RunLog.Info("node detail db query success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: extResp}
}

func setNodeExtInfos(nodeInfo NodeInfoDetail) (NodeInfoDetail, error) {
	var err error
	nodeInfo.NodeGroup, err = evalNodeGroup(nodeInfo.ID)
	if err != nil {
		hwlog.RunLog.Errorf("get node detail by sn db query error, %s", err.Error())
		return nodeInfo, err
	}

	nodeResource, err := NodeSyncInstance().GetAllocatableResource(nodeInfo.UniqueName)
	if err != nil {
		hwlog.RunLog.Warnf("get node detail query node resource error, %s", err.Error())
		nodeResource = &NodeResource{}
	}
	nodeInfo.NodeResourceInfo = NodeResourceInfo{
		Cpu:    nodeResource.Cpu.Value(),
		Memory: nodeResource.Memory.Value(),
		Npu:    nodeResource.Npu.Value(),
	}
	nodeInfo.Status, err = NodeSyncInstance().GetMEFNodeStatus(nodeInfo.UniqueName)
	if err != nil {
		hwlog.RunLog.Warnf("get node detail query node status error, %s", err.Error())
		nodeInfo.Status = statusOffline
	}

	return nodeInfo, nil
}

func modifyNode(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start modify node")
	var req ModifyNodeReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("modify node convert request error, %s", err.Error())
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}
	if checkResult := newModifyNodeChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("modify node check parameters failed, %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}
	updatedColumns := map[string]interface{}{
		"NodeName":  req.NodeName,
		"UpdatedAt": time.Now().Format(TimeFormat),
	}
	if req.Description != nil {
		updatedColumns["Description"] = req.Description
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

func getNodeStatistics(*model.Message) common.RespMsg {
	hwlog.RunLog.Info("start get node statistics")
	nodes, err := NodeServiceInstance().listNodes()
	if err != nil {
		hwlog.RunLog.Error("failed to get node statistics, db query failed")
		return common.RespMsg{Status: common.ErrorCountNodeByStatus, Msg: ""}
	}
	statusMap := make(map[string]string)
	allNodeStatus := NodeSyncInstance().ListMEFNodeStatus()
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

func listManagedNode(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start list node managed")
	var req types.ListReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("list node parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}
	if checkResult := util.NewPaginationQueryChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("list node para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}
	total, err := NodeServiceInstance().countNodesByName(req.Name, managed)
	if err != nil {
		hwlog.RunLog.Error("count node failed")
		return common.RespMsg{Status: common.ErrorListNode, Msg: "count node failed", Data: nil}
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
		respItem.Status, err = NodeSyncInstance().GetMEFNodeStatus(nodeInfo.UniqueName)
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

func listUnmanagedNode(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start list node unmanaged")
	var req types.ListReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("list node convert request error: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}
	if checkResult := util.NewPaginationQueryChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("list unmanaged node para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
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
			respItem.Status, err = NodeSyncInstance().GetMEFNodeStatus(node.UniqueName)
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
	hwlog.RunLog.Error("list unmanaged node failed")
	return common.RespMsg{Status: common.ErrorListUnManagedNode, Msg: "", Data: nil}
}

func listNode(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start list all nodes")
	var req types.ListReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("list nodes parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}
	if checkResult := util.NewPaginationQueryChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("list nodes para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}
	total, err := NodeServiceInstance().countAllNodesByName(req.Name)
	if err != nil {
		hwlog.RunLog.Error("count node failed")
		return common.RespMsg{Status: common.ErrorListNode, Msg: "get node total num failed", Data: nil}
	}
	nodes, err := NodeServiceInstance().listAllNodesByName(req.PageNum, req.PageSize, req.Name)
	if err != nil {
		hwlog.RunLog.Error("list all nodes failed")
		return common.RespMsg{Status: common.ErrorListNode, Msg: "list all nodes failed", Data: nil}
	}
	var nodeList []NodeInfoExManaged
	for _, nodeInfo := range *nodes {
		var respItem NodeInfoExManaged
		respItem.NodeGroup, err = evalNodeGroup(nodeInfo.ID)
		if err != nil {
			hwlog.RunLog.Errorf("list node id [%d] db error: %s", nodeInfo.ID, err.Error())
			return common.RespMsg{Status: common.ErrorListNode, Msg: err.Error()}
		}
		respItem.Status, err = NodeSyncInstance().GetMEFNodeStatus(nodeInfo.UniqueName)
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

func batchDeleteNode(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start delete node")
	var req BatchDeleteNodeReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("failed to delete node, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed"}
	}
	if checkResult := newBatchDeleteNodeChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("failed to delete node, error: %v", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}
	var res types.BatchResp
	var successNodes []interface{}
	failedMap := make(map[string]string)
	res.FailedInfos = failedMap
	for _, nodeID := range req.NodeIDs {
		nodeInfo, err := deleteSingleNode(nodeID)
		if err != nil {
			if nodeInfo == nil {
				hwlog.RunLog.Errorf("failed to delete node %d, error: err=%v", nodeID, err)
			} else {
				hwlog.RunLog.Errorf("failed to delete node %d(sn=%s), error: err=%v",
					nodeID, nodeInfo.SerialNumber, err)
			}
			failedMap[strconv.Itoa(int(nodeID))] = fmt.Sprintf("failed to delete, error: err=%v", err)
			continue
		}
		res.SuccessIDs = append(res.SuccessIDs, nodeID)
		successNodes = append(successNodes, nodeInfo.SerialNumber)
	}
	logmgmt.BatchOperationLog("batch delete node with sn", successNodes)
	if len(res.FailedInfos) != 0 {
		return common.RespMsg{Status: common.ErrorDeleteNode, Data: res}
	}
	hwlog.RunLog.Info("delete node success")
	return common.RespMsg{Status: common.Success, Data: nil}
}

func deleteSingleNode(nodeID uint64) (*NodeInfo, error) {
	nodeInfo, err := NodeServiceInstance().getNodeByID(nodeID)
	if err != nil {
		return nil, errors.New("db query failed")
	}
	if !nodeInfo.IsManaged {
		return nodeInfo, errors.New("can't delete unmanaged node")
	}

	nodeRelations, err := NodeServiceInstance().getRelationsByNodeID(nodeID)
	if err != nil {
		hwlog.RunLog.Errorf("query node(%d) group info failed, error: %v", nodeID, err)
		return nodeInfo, fmt.Errorf("query node(%d) group info failed", nodeID)
	}
	for _, relation := range *nodeRelations {
		count, err := getAppInstanceCountByGroupId(relation.GroupID)
		if err != nil {
			return nodeInfo, fmt.Errorf("query group(%d) app count failed, %v", relation.GroupID, err)
		}
		if count > 0 {
			return nodeInfo, fmt.Errorf("group(%d) has deployed app, can't remove", relation.GroupID)
		}
	}

	if err = NodeServiceInstance().deleteNode(nodeInfo); err != nil {
		return nodeInfo, err
	}
	if err := sendDeleteNodeMessageToNode(nodeInfo.SerialNumber); err != nil {
		return nodeInfo, fmt.Errorf("send delete node msg error:%v", err)
	}
	return nodeInfo, nil
}

func deleteNodeFromGroup(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start delete node from group")
	var req DeleteNodeToGroupReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("failed to delete from group, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed"}
	}
	if checkResult := newDeleteNodeFromGroupChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("failed to delete node from group, error: %v", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}
	var res types.BatchResp
	var successNodes []interface{}
	var groupInfo *NodeGroup
	failedMap := make(map[string]string)
	res.FailedInfos = failedMap
	groupInfo, err := NodeServiceInstance().getNodeGroupByID(*req.GroupID)
	if err != nil {
		if groupInfo == nil {
			hwlog.RunLog.Errorf("failed to get group %d's info, error: ret is empty", *req.GroupID)
		} else {
			hwlog.RunLog.Errorf("failed to get group %d's info, error: %v", *req.GroupID, err)
		}
		for _, nodeID := range *req.NodeIDs {
			failedMap[strconv.Itoa(int(nodeID))] = "failed to delete, error: failed to get group info"
		}
	} else {
		for _, nodeID := range *req.NodeIDs {
			nodeInfo, err := NodeServiceInstance().getNodeByID(nodeID)
			if err != nil {
				hwlog.RunLog.Errorf("failed to get node %d's info, error: %v", nodeID, err)
				failedMap[strconv.Itoa(int(nodeID))] = "failed to delete, error: failed to get node info"
				continue
			}
			if err = NodeServiceInstance().deleteSingleNodeRelation(*req.GroupID, nodeID); err != nil {
				hwlog.RunLog.Errorf("failed to delete node %d(sn=%s) from group %d(name=%s), error: err=%v", nodeID,
					nodeInfo.SerialNumber, *req.GroupID, groupInfo.GroupName, err)
				failedMap[strconv.Itoa(int(nodeID))] = fmt.Sprintf("failed to delete, error: err=%v", err)
				continue
			}
			res.SuccessIDs = append(res.SuccessIDs, nodeID)
			successNodes = append(successNodes, nodeInfo.SerialNumber)
		}
		logmgmt.BatchOperationLog(fmt.Sprintf("from group [%s], delete node", groupInfo.GroupName), successNodes)
	}
	if len(res.FailedInfos) != 0 {
		return common.RespMsg{Status: common.ErrorDeleteNodeFromGroup, Msg: "", Data: res}
	}
	hwlog.RunLog.Info("delete node relation success")
	return common.RespMsg{Status: common.Success, Data: nil}
}

func batchDeleteNodeRelation(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start delete node relation")
	var req BatchDeleteNodeRelationReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("failed to delete node relation, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed"}
	}
	if checkResult := newBatchDeleteNodeRelationChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("failed to delete node relation, error: %v", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}
	var res types.BatchResp
	var successRelations []interface{}
	failedMap := make(map[string]string)
	res.FailedInfos = failedMap
	for _, relation := range req {
		relationStr := fmt.Sprintf("groupID: %d, nodeID: %d", *relation.GroupID, *relation.NodeID)
		nodeSn, groupName, err := deleteSingleRelation(*relation.GroupID, *relation.NodeID)
		if err != nil {
			failedMap[relationStr] = err.Error()
			continue
		}
		res.SuccessIDs = append(res.SuccessIDs, relation)
		successRelations = append(successRelations, DeleteNodeRelationRecord{
			NodeSn:    nodeSn,
			GroupName: groupName,
		})
	}
	logmgmt.BatchOperationLog("batch delete node relation", successRelations)
	if len(res.FailedInfos) != 0 {
		return common.RespMsg{Status: common.ErrorDeleteNodeFromGroup, Msg: "", Data: res}
	}
	hwlog.RunLog.Info("delete node relation success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func deleteSingleRelation(groupId, nodeId uint64) (string, string, error) {
	nodeInfo, err := NodeServiceInstance().getNodeByID(nodeId)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get node %d's info, error: %v", nodeId, err)
		return "", "", errors.New("failed to delete node relation, error: failed to get node info")
	}
	if nodeInfo == nil {
		hwlog.RunLog.Errorf("failed to get node %d's info, error: ret is empty", nodeId)
		return "", "", errors.New("failed to delete node relation, error: failed to get node info")
	}
	groupInfo, err := NodeServiceInstance().getNodeGroupByID(groupId)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get group %d's info, error: %v", groupId, err)
		return "", "", errors.New("failed to delete node relation: failed to get group info")
	}
	if groupInfo == nil {
		hwlog.RunLog.Errorf("failed to get group %d's info, error: ret is empty", groupId)
		return "", "", errors.New("failed to delete node relation: failed to get group info")
	}
	if err = NodeServiceInstance().deleteSingleNodeRelation(groupId, nodeId); err != nil {
		hwlog.RunLog.Errorf("failed to delete node relation: node %d(sn=%s) from group %d(name=%s), error: %v",
			nodeId, nodeInfo.SerialNumber, groupId, groupInfo.GroupName, err)
		return "", "", fmt.Errorf("failed to delete node relation, error: %v", err)
	}

	return nodeInfo.SerialNumber, groupInfo.GroupName, nil
}

func isNodeNotFound(err error) bool {
	return nodeNotFoundPattern.MatchString(err.Error())
}

func evalNodeGroup(nodeID uint64) (string, error) {
	relations, err := NodeServiceInstance().getRelationsByNodeID(nodeID)
	if err != nil {
		hwlog.RunLog.Error("get node detail db query error")
		return "", errors.New("get node relation by id failed")
	}
	nodeGroupName := ""
	if len(*relations) > 0 {
		var buffer bytes.Buffer
		for index, relation := range *relations {
			nodeGroup, err := NodeServiceInstance().getNodeGroupByID(relation.GroupID)
			if err != nil {
				hwlog.RunLog.Error("get node group by id failed")
				return "", errors.New("get node group by id failed")
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

func addNodeRelation(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start add node to group")
	var req AddNodeToGroupReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("add node to group convert request error, %s", err.Error())
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}
	if checkResult := newAddNodeRelationChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("failed to add node to group, error: %v", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}

	nodeServiceChecker := specificationChecker{nodeService: NodeServiceInstance()}
	if err := nodeServiceChecker.checkAddNodeToGroup(*req.NodeIDs, []uint64{*req.GroupID}); err != nil {
		hwlog.RunLog.Errorf("add node to group check spec error: %s", err.Error())
		return common.RespMsg{Status: common.ErrorCheckNodeMrgSize, Msg: err.Error()}
	}
	res, err := addNode(req)
	if err != nil {
		hwlog.RunLog.Error("add node to group failed")
		return common.RespMsg{Status: common.ErrorAddNodeToGroup, Msg: err.Error(), Data: res}
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func addNode(req AddNodeToGroupReq) (*types.BatchResp, error) {
	var res types.BatchResp
	var successNodes []interface{}
	failedMap := make(map[string]string)
	res.FailedInfos = failedMap

	nodeGroup, err := NodeServiceInstance().getNodeGroupByID(*req.GroupID)
	if err != nil {
		hwlog.RunLog.Errorf("add node failed, dont have this node group id(%d)", *req.GroupID)
		return nil, fmt.Errorf("dont have this node group id(%d)", *req.GroupID)
	}
	resReq, count, err := getRequestItemsOfAddGroup(nodeGroup)
	if err != nil {
		hwlog.RunLog.Errorf("get group id [%d] request items for add to group failed, %v", nodeGroup.ID, err)
		return nil, fmt.Errorf("get group id [%d] request items for add to group failed, %v", nodeGroup.ID, err)
	}
	for i, id := range *req.NodeIDs {
		if err = checkNodeBeforeAddToGroup(resReq, count, id); err != nil {
			hwlog.RunLog.Errorf("add node[%d] to group[%d](name=%s) error, check node failed: %s",
				id, nodeGroup.ID, nodeGroup.GroupName, err.Error())
			failedMap[strconv.Itoa(int(id))] = fmt.Sprintf("add node to group error, check node failed: %s",
				err.Error())
			continue
		}
		nodeDb, err := NodeServiceInstance().getManagedNodeByID(id)
		if err != nil {
			hwlog.RunLog.Errorf("does not find node id %d", id)
			failedMap[strconv.Itoa(int(id))] = "does not find the node"
			continue
		}
		relation := NodeRelation{
			NodeID:    (*req.NodeIDs)[i],
			GroupID:   *req.GroupID,
			CreatedAt: time.Now().Format(TimeFormat)}
		if err = nodeServiceInstance.addNodeToGroup(&relation, nodeDb.UniqueName); err != nil {
			hwlog.RunLog.Errorf("add node[%s](sn=%s) to group[%d](name=%s) error: %v",
				nodeDb.NodeName, nodeDb.SerialNumber, nodeGroup.ID, nodeGroup.GroupName, err)
			failedMap[strconv.Itoa(int(id))] = "add node to group failed"
			continue
		}
		res.SuccessIDs = append(res.SuccessIDs, id)
		successNodes = append(successNodes, nodeDb.SerialNumber)
	}
	logmgmt.BatchOperationLog(fmt.Sprintf("into group [%s], add node", nodeGroup.GroupName), successNodes)
	if len(res.FailedInfos) != 0 {
		return &res, errors.New("add some nodes to group failed")
	}
	return &res, nil
}

func getRequestItemsOfAddGroup(nodeGroup *NodeGroup) (v1.ResourceList, int64, error) {
	resReq, err := getNodeGroupResReq(nodeGroup)
	if err != nil {
		return nil, 0, err
	}
	count, err := getAppInstanceCountByGroupId(nodeGroup.ID)
	if err != nil {
		return nil, 0, fmt.Errorf("get node group id [%d] deployed app count request error", nodeGroup.ID)
	}
	return resReq, count, nil
}

func addUnManagedNode(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start add unmanaged node")
	var req AddUnManagedNodeReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("add unmanaged node convert request error, %s", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed", Data: nil}
	}
	if checkResult := newAddUnManagedNodeChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("add unmanaged node validate parameters error, %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}

	nodeServiceChecker := specificationChecker{nodeService: NodeServiceInstance()}
	if err := nodeServiceChecker.checkAddNodeToGroup([]uint64{*req.NodeID}, req.GroupIDs); err != nil {
		hwlog.RunLog.Errorf("add unmanaged node to group check spec error: %s", err.Error())
		return common.RespMsg{Status: common.ErrorCheckNodeMrgSize, Msg: err.Error()}
	}
	err := NodeServiceInstance().checkNodeManagedStatus(*req.NodeID, unmanaged)
	if err != nil {
		hwlog.RunLog.Error(err.Error())
		return common.RespMsg{Status: common.ErrorAddUnManagedNode, Msg: err.Error(), Data: nil}
	}
	updatedColumns := map[string]interface{}{
		"NodeName":    req.NodeName,
		"Description": req.Description,
		"IsManaged":   managed,
		"UpdatedAt":   time.Now().Format(TimeFormat),
	}
	if cnt, err := NodeServiceInstance().updateNode(*req.NodeID, unmanaged, updatedColumns); err != nil || cnt != 1 {
		hwlog.RunLog.Errorf("add unmanaged node error: %v", err)
		return common.RespMsg{Status: common.ErrorAddUnManagedNode, Msg: "add node to mef system error", Data: nil}
	}
	addNodeRes := addNodeToGroups(req)
	if len(addNodeRes.FailedInfos) != 0 {
		return common.RespMsg{Status: common.ErrorAddUnManagedNode,
			Msg: "add node to mef success, but node cannot join some group", Data: addNodeRes}
	}
	hwlog.RunLog.Info("add unmanaged node success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func addNodeToGroups(req AddUnManagedNodeReq) types.BatchResp {
	var addNodeRes types.BatchResp
	failedMap := make(map[string]string)
	addNodeRes.FailedInfos = failedMap
	for _, id := range req.GroupIDs {
		addReq := AddNodeToGroupReq{NodeIDs: &[]uint64{*req.NodeID}, GroupID: &id}
		_, err := addNode(addReq)
		if err != nil {
			failedMap[strconv.Itoa(int(id))] = err.Error()
			continue
		}
		addNodeRes.SuccessIDs = append(addNodeRes.SuccessIDs, id)
	}
	return addNodeRes
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

func getNodeGroupResReq(nodeGroup *NodeGroup) (v1.ResourceList, error) {
	var allocatedRes v1.ResourceList
	if nodeGroup.ResourcesRequest == "" {
		return make(map[v1.ResourceName]resource.Quantity), nil
	}
	err := json.Unmarshal([]byte(nodeGroup.ResourcesRequest), &allocatedRes)
	if err != nil {
		return nil, errors.New("unmarshal node group resource request error")
	}
	return allocatedRes, nil
}

func checkNodeResource(req v1.ResourceList, nodeId uint64) error {
	nodeInfo, err := NodeServiceInstance().getNodeByID(nodeId)
	if err != nil {
		hwlog.RunLog.Errorf("get node info by node id [%d] error: %v", nodeId, err)
		return fmt.Errorf("get node info by node id [%d] error", nodeId)
	}
	availableRes, err := NodeSyncInstance().GetAvailableResource(nodeInfo.ID, nodeInfo.UniqueName)
	if err != nil {
		return fmt.Errorf("get node allocatable resource by node unique name [%s] error: %s",
			nodeInfo.UniqueName, err.Error())
	}

	if availableRes.Cpu.Cmp(*req.Cpu()) < 0 {
		return fmt.Errorf("node [%d] do not have enough cpu resources", nodeId)
	}
	if availableRes.Memory.Cmp(*req.Memory()) < 0 {
		return fmt.Errorf("node [%d] do not have enough memory resources", nodeId)
	}
	npuReq, ok := req[common.DeviceType]
	if ok && availableRes.Npu.Cmp(npuReq) < 0 {
		return fmt.Errorf("node [%d] do not have enough npu resources", nodeId)
	}
	return nil
}

func checkNodePodLimit(addedNumber int64, nodeId uint64) error {
	nodeGroups, err := NodeServiceInstance().getGroupsByNodeID(nodeId)
	if err != nil {
		return fmt.Errorf("get node groups by node id [%d] error", nodeId)
	}
	var nodeDeployedCount int64
	for _, group := range *nodeGroups {
		count, err := getAppInstanceCountByGroupId(group.ID)
		if err != nil {
			return fmt.Errorf("get deployed app count by node group id [%d] error", group.ID)
		}
		nodeDeployedCount += count
	}
	if addedNumber+nodeDeployedCount > config.PodConfig.MaxPodNumberPerNode {
		return fmt.Errorf("pod addedNumber is out of node [%d] max allowed addedNumber", nodeId)
	}
	return nil
}

func checkNodeBeforeAddToGroup(req v1.ResourceList, addedNumber int64, nodeId uint64) error {
	if err := checkNodeResource(req, nodeId); err != nil {
		return err
	}
	if err := checkNodePodLimit(addedNumber, nodeId); err != nil {
		return err
	}
	return nil
}

func updateNodeSoftwareInfo(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start to update node software info")
	const maxPayloadSize = 1024

	var inputStr string
	if err := msg.ParseContent(&inputStr); err != nil {
		hwlog.RunLog.Errorf("update node software info failed: parse content failed: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed"}
	}

	if len(inputStr) > maxPayloadSize {
		hwlog.RunLog.Error("software info size exceeded")
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: "software info size exceeded", Data: nil}
	}

	var req types.EdgeReportSoftwareInfoReq
	if err := common.ParamConvert(inputStr, &req); err != nil {
		hwlog.RunLog.Errorf("update node software info error, %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "convert request error", Data: nil}
	}

	softwareInfo, err := json.Marshal(req.SoftwareInfo)
	if err != nil {
		hwlog.RunLog.Error("marshal version info failed")
		return common.RespMsg{Status: "", Msg: "marshal version info failed", Data: nil}
	}

	nodeInfo, err := NodeServiceInstance().getNodeInfoBySerialNumber(msg.GetPeerInfo().Sn)
	if err != nil {
		hwlog.RunLog.Errorf("get node info [%s] failed:%v", msg.GetPeerInfo().Sn, err)
		return common.RespMsg{Status: "", Msg: "get node info failed", Data: nil}
	}

	count, err := GetTableCount(NodeInfo{})
	if err != nil && err != gorm.ErrRecordNotFound {
		hwlog.RunLog.Errorf("get nodes count failed: %v", err)
		return common.RespMsg{Status: "", Msg: "get node info failed", Data: nil}
	}

	if count > maxNodeInfos {
		hwlog.RunLog.Error("nodes count exceeds the limitation")
		return common.RespMsg{Status: "", Msg: "nodes count exceeds the limitation", Data: nil}
	}

	nodeInfo.SoftwareInfo = string(softwareInfo)
	err = NodeServiceInstance().updateNodeInfoBySerialNumber(msg.GetPeerInfo().Sn, nodeInfo)
	if err != nil {
		hwlog.RunLog.Errorf("update node software info failed: %v", err)
		return common.RespMsg{Status: "", Msg: "update node software info failed", Data: nil}
	}

	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func deleteUnManagedNode(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start delete unmanaged node")
	var req BatchDeleteNodeReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("failed to delete unmanaged node, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed"}
	}
	if checkResult := newBatchDeleteNodeChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("failed to delete unmanaged node, error: %v", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}
	var res types.BatchResp
	var successNodes []interface{}
	failedMap := make(map[string]string)
	res.FailedInfos = failedMap
	for _, nodeID := range req.NodeIDs {
		nodeInfo, err := deleteSingleUnManagedNode(nodeID)
		if err != nil {
			if nodeInfo == nil {
				hwlog.RunLog.Errorf("failed to delete unmanaged node %d, error: err=%v", nodeID, err)
			} else {
				hwlog.RunLog.Errorf("failed to delete unmanaged node %d(sn=%s), error: err=%v",
					nodeID, nodeInfo.SerialNumber, err)
			}
			failedMap[strconv.Itoa(int(nodeID))] = fmt.Sprintf("failed to delete, error: err=%v", err)
			continue
		}
		res.SuccessIDs = append(res.SuccessIDs, nodeID)
		successNodes = append(successNodes, nodeInfo.SerialNumber)
	}
	logmgmt.BatchOperationLog("batch delete unmanaged node", successNodes)
	if len(res.FailedInfos) != 0 {
		return common.RespMsg{Status: common.ErrorDeleteNode, Data: res}
	}
	hwlog.RunLog.Info("delete unmanaged node success")
	return common.RespMsg{Status: common.Success}
}

func sendDeleteNodeMessageToNode(serialNumber string) error {
	sendMsg, err := model.NewMessage()
	if err != nil {
		return fmt.Errorf("create new message failed, error: %v", err)
	}
	sendMsg.SetNodeId(serialNumber)
	sendMsg.SetRouter(common.NodeManagerName, common.CloudHubName, common.Delete, common.DeleteNodeMsg)
	if err = sendMsg.FillContent(fmt.Sprintf("delete:%s", serialNumber)); err != nil {
		return fmt.Errorf("fill content failed: %v", err)
	}
	if err = modulemgr.SendMessage(sendMsg); err != nil {
		return fmt.Errorf("%s sends message to %s failed, error: %v",
			common.NodeManagerName, common.CloudHubName, err)
	}
	return nil
}

func deleteSingleUnManagedNode(nodeID uint64) (*NodeInfo, error) {
	nodeInfo, err := NodeServiceInstance().getNodeByID(nodeID)
	if err != nil {
		return nil, errors.New("db query failed")
	}
	if nodeInfo.IsManaged {
		return nodeInfo, errors.New("node is managed")
	}
	if err = NodeServiceInstance().deleteUnmanagedNode(nodeInfo); err != nil {
		return nodeInfo, err
	}
	if err := sendDeleteNodeMessageToNode(nodeInfo.SerialNumber); err != nil {
		return nodeInfo, fmt.Errorf("send delete node msg error:%v", err)
	}
	return nodeInfo, nil
}
