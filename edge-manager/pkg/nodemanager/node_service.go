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
	"huawei.com/mindx/common/hwlog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"edge-manager/pkg/types"
	"edge-manager/pkg/util"

	"huawei.com/mindxedge/base/common"
)

var (
	nodeNotFoundPattern = regexp.MustCompile(`nodes "([^"]+)" not found`)
)

const groupLabelLen = 4

// getNodeDetail get node detail
func getNodeDetail(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start get node detail")
	id, ok := input.(uint64)
	if !ok {
		hwlog.RunLog.Error("query node detail failed: para type not valid")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "query node detail convert param failed"}
	}
	if checkResult := newGetNodeDetailChecker().Check(id); !checkResult.Result {
		hwlog.RunLog.Errorf("query node detail parameters failed, %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
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
	nodeResource, err := NodeSyncInstance().GetAllocatableResource(nodeInfo.UniqueName)
	if err != nil {
		hwlog.RunLog.Warnf("get node detail query node resource error, %s", err.Error())
		nodeResource = &NodeResource{}
	}
	resp.NodeResourceInfo = NodeResourceInfo{
		Cpu:    nodeResource.Cpu.Value(),
		Memory: nodeResource.Memory.Value(),
		Npu:    nodeResource.Npu.Value(),
	}
	resp.Status, err = NodeSyncInstance().GetNodeStatus(nodeInfo.UniqueName)
	if err != nil {
		hwlog.RunLog.Warnf("get node detail query node status error, %s", err.Error())
		resp.Status = statusOffline
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
	allNodeStatus := NodeSyncInstance().ListNodeStatus()
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
		respItem.Status, err = NodeSyncInstance().GetNodeStatus(nodeInfo.UniqueName)
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
	hwlog.RunLog.Info("start list node unmanaged")
	req, ok := input.(types.ListReq)
	if !ok {
		hwlog.RunLog.Error("list node convert request error")
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "", Data: nil}
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
			respItem.Status, err = NodeSyncInstance().GetNodeStatus(node.UniqueName)
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

func listNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start list all nodes")
	req, ok := input.(types.ListReq)
	if !ok {
		hwlog.RunLog.Error("list nodes convert request error")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "convert request error", Data: nil}
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
		respItem.Status, err = NodeSyncInstance().GetNodeStatus(nodeInfo.UniqueName)
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
	failedMap := make(map[string]string)
	res.FailedInfos = failedMap
	for _, nodeID := range req.NodeIDs {
		if err := deleteSingleNode(nodeID); err != nil {
			errInfo := fmt.Sprintf("failed to delete node, error: err=%v", err)
			hwlog.RunLog.Error(errInfo)
			failedMap[strconv.Itoa(int(nodeID))] = errInfo
			continue
		}
		res.SuccessIDs = append(res.SuccessIDs, nodeID)
	}
	if len(res.FailedInfos) != 0 {
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

	nodeRelations, err := NodeServiceInstance().getRelationsByNodeID(nodeID)
	if err != nil {
		return err
	}
	for _, relation := range *nodeRelations {
		count, err := getAppInstanceCountByGroupId(relation.GroupID)
		if err != nil {
			return fmt.Errorf("query group(%d) app count failed, %v", relation.GroupID, err)
		}
		if count > 0 {
			return fmt.Errorf("group(%d) has deployed app, can't remove", relation.GroupID)
		}
	}

	if err = NodeServiceInstance().deleteNode(nodeInfo); err != nil {
		return err
	}
	return nil
}

func deleteNodeFromGroup(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start delete node from group")
	var req DeleteNodeToGroupReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("failed to delete from group, error: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error()}
	}
	if checkResult := newDeleteNodeFromGroupChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("failed to delete node from group, error: %v", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}
	var res types.BatchResp
	failedMap := make(map[string]string)
	res.FailedInfos = failedMap
	for _, nodeID := range *req.NodeIDs {
		if err := NodeServiceInstance().deleteSingleNodeRelation(*req.GroupID, nodeID); err != nil {
			errInfo := fmt.Sprintf("failed to delete node from group, error: err=%v", err)
			hwlog.RunLog.Error(errInfo)
			failedMap[strconv.Itoa(int(nodeID))] = errInfo
			continue
		}
		res.SuccessIDs = append(res.SuccessIDs, nodeID)
	}
	if len(res.FailedInfos) != 0 {
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
	failedMap := make(map[string]string)
	res.FailedInfos = failedMap
	for _, relation := range req {
		if err := NodeServiceInstance().deleteSingleNodeRelation(*relation.GroupID, *relation.NodeID); err != nil {
			errInfo := fmt.Sprintf("failed to delete node relation, error: %v", err)
			relationStr := fmt.Sprintf("groupID: %d, nodeID: %d", *relation.GroupID, *relation.NodeID)
			hwlog.RunLog.Error(errInfo)
			failedMap[relationStr] = errInfo
			continue
		}
		res.SuccessIDs = append(res.SuccessIDs, relation)
	}
	if len(res.FailedInfos) != 0 {
		return common.RespMsg{Status: common.ErrorDeleteNodeFromGroup, Msg: "", Data: res}
	}
	hwlog.RunLog.Info("delete node relation success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
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
	failedMap := make(map[string]string)
	res.FailedInfos = failedMap

	nodeGroup, err := NodeServiceInstance().getNodeGroupByID(*req.GroupID)
	if err != nil {
		return nil, fmt.Errorf("dont have this node group id(%d)", *req.GroupID)
	}
	resReq, err := getNodeGroupResReq(nodeGroup)
	if err != nil {
		return nil, fmt.Errorf("parse node group id [%d] resources request error", *req.GroupID)
	}
	for i, id := range *req.NodeIDs {
		if err = checkNodeResource(resReq, id); err != nil {
			errInfo := fmt.Sprintf("check node allocatable resource failed: %s", err.Error())
			hwlog.RunLog.Error(errInfo)
			failedMap[strconv.Itoa(int(id))] = errInfo
			continue
		}
		nodeDb, err := NodeServiceInstance().getManagedNodeByID(id)
		if err != nil {
			errInfo := fmt.Sprintf("no found node id %d", id)
			hwlog.RunLog.Error(errInfo)
			failedMap[strconv.Itoa(int(id))] = errInfo
			continue
		}
		relation := NodeRelation{
			NodeID:    (*req.NodeIDs)[i],
			GroupID:   *req.GroupID,
			CreatedAt: time.Now().Format(TimeFormat)}
		if err := nodeServiceInstance.addNodeToGroup(&relation, nodeDb.UniqueName); err != nil {
			errInfo := fmt.Sprintf("add node(%s) to group(%d) error", nodeDb.NodeName, nodeGroup.ID)
			hwlog.RunLog.Error(errInfo)
			failedMap[strconv.Itoa(int(id))] = errInfo
			continue
		}
		res.SuccessIDs = append(res.SuccessIDs, id)
	}
	if len(res.FailedInfos) != 0 {
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
	if len(addNodeRes.FailedInfos) != 0 {
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
		return fmt.Errorf("get node info by node id [%d] error", nodeId)
	}
	availableRes, err := NodeSyncInstance().GetAvailableResource(nodeInfo.UniqueName)
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

func updateNodeSoftwareInfo(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start to update node software info")
	var req types.EdgeReportSoftwareInfoReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("update node software info error, %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "convert request error", Data: nil}
	}

	softwareInfo, err := json.Marshal(req.SoftwareInfo)
	if err != nil {
		hwlog.RunLog.Error("marshal version info failed")
		return common.RespMsg{Status: "", Msg: "marshal version info failed", Data: nil}
	}

	nodeInfo, err := NodeServiceInstance().getNodeInfoBySerialNumber(req.SerialNumber)
	if err != nil && err != gorm.ErrRecordNotFound {
		hwlog.RunLog.Errorf("get node info [%s] failed:%v", req.SerialNumber, err)
		return common.RespMsg{Status: "", Msg: "get node info failed", Data: nil}
	}

	nodeInfo.SoftwareInfo = string(softwareInfo)
	err = NodeServiceInstance().updateNodeInfoBySerialNumber(req.SerialNumber, nodeInfo)
	if err != nil {
		hwlog.RunLog.Errorf("update node software info failed: %v", err)
		return common.RespMsg{Status: "", Msg: "update node software info failed", Data: nil}
	}

	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}
