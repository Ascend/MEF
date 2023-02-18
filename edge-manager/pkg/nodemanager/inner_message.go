// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package nodemanager

import (
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"k8s.io/api/core/v1"

	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
)

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

func innerGetNodesByNodeGroupID(input interface{}) common.RespMsg {
	req, ok := input.(types.InnerGetNodesReq)
	if !ok {
		hwlog.RunLog.Error("parse inner message content failed")
		return common.RespMsg{Status: "", Msg: "parse inner message content failed"}
	}
	relations, err := NodeServiceInstance().listNodeRelationsByGroupId(req.NodeGroupID)
	if err != nil {
		hwlog.RunLog.Error("inner message get node id failed")
		return common.RespMsg{Status: "", Msg: "inner message get node status failed"}
	}
	var nodeIDs []uint64
	for _, relation := range *relations {
		nodeIDs = append(nodeIDs, relation.NodeID)
	}
	resp := types.InnerGetNodesResp{
		NodeIDs: nodeIDs,
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

func innerAllNodeInfos(input interface{}) common.RespMsg {
	_, ok := input.(types.InnerGetNodeInfoResReq)
	if !ok {
		hwlog.RunLog.Error("parse inner message content failed")
		return common.RespMsg{Status: "", Msg: "parse inner message content failed"}
	}
	nodeInfos, err := NodeServiceInstance().listNodes()
	if err != nil {
		hwlog.RunLog.Error("inner message get all node info failed")
		return common.RespMsg{Status: "", Msg: "internal get all node info failed", Data: nil}
	}
	hwlog.RunLog.Info("inner message get all node info success")
	return common.RespMsg{Status: common.Success, Msg: "internal get all node info success", Data: nodeInfos}
}

func innerCheckNodeGroupResReq(input interface{}) common.RespMsg {
	req, ok := input.(types.InnerCheckNodeResReq)
	if !ok {
		hwlog.RunLog.Error("parse inner message content failed")
		return common.RespMsg{Status: "", Msg: "parse inner message content failed"}
	}
	if err := checkNodeResBeforeDeployApp(req); err != nil {
		hwlog.RunLog.Errorf("check node allocated resource before deploying app failed: %s", err.Error())
		return common.RespMsg{Status: "", Msg: err.Error()}
	}
	return common.RespMsg{Status: common.Success}
}

func checkNodeResBeforeDeployApp(req types.InnerCheckNodeResReq) error {
	nodeRelations, err := NodeServiceInstance().listNodeRelationsByGroupId(req.NodeGroupID)
	if err != nil {
		return fmt.Errorf("get node relations by group id [%d] error", req.NodeGroupID)
	}
	for _, nodeRelation := range *nodeRelations {
		if err := checkNodeResource(req.ResourceReqs, nodeRelation.NodeID); err != nil {
			return fmt.Errorf("in group [%d], %s", req.NodeGroupID, err.Error())
		}
	}
	return nil
}

func innerUpdateNodeGroupResReq(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start to update node group allocated resources")
	req, ok := input.(types.InnerUpdateNodeResReq)
	if !ok {
		hwlog.RunLog.Error("parse inner message content failed")
		return common.RespMsg{Status: "", Msg: "parse inner message content failed"}
	}
	err := updateNodeGroupResReq(req.ResourceReqs, req.NodeGroupID, req.IsUndeploy)
	if err != nil {
		hwlog.RunLog.Errorf("update node group resource request failed: %s", err.Error())
		return common.RespMsg{Status: "", Msg: err.Error(), Data: false}
	}
	return common.RespMsg{Status: common.Success}
}

func updateNodeGroupResReq(req v1.ResourceList, nodeGroupID uint64, isUndeploy bool) error {
	nodeGroup, err := NodeServiceInstance().getNodeGroupByID(nodeGroupID)
	if err != nil {
		return fmt.Errorf("get node group id [%d] resources request failed, db error", nodeGroupID)
	}
	allocatedRes, err := getNodeGroupResReq(nodeGroup)
	if err != nil {
		return fmt.Errorf("parse node group id [%d] resources request error", nodeGroupID)
	}
	for name, quantity := range req {
		currentRes := allocatedRes[name]
		if isUndeploy {
			currentRes.Sub(quantity)
		} else {
			currentRes.Add(quantity)
		}
		allocatedRes[name] = currentRes
	}
	accumRes, err := json.Marshal(allocatedRes)
	if err != nil {
		return fmt.Errorf("marshal node group id [%d] resources request error", nodeGroupID)
	}
	updatedColumns := map[string]interface{}{"ResourcesRequest": accumRes}
	if cnt, err := NodeServiceInstance().updateNodeGroupRes(nodeGroupID, updatedColumns); err != nil || cnt != 1 {
		return fmt.Errorf("update resources request to node group id [%d] failed, db error", nodeGroupID)
	}
	return nil
}

func innerGetNodeGroupInfosByIds(input interface{}) common.RespMsg {
	req, ok := input.(types.InnerGetNodeGroupInfosReq)
	if !ok {
		hwlog.RunLog.Error("parse inner message content failed")
		return common.RespMsg{Status: "", Msg: "parse inner message content failed", Data: nil}
	}
	var nodeGroupInfos []types.NodeGroupInfo
	for _, id := range req.NodeGroupIds {
		nodeGroupInfo, err := NodeServiceInstance().getNodeGroupByID(id)
		if err == gorm.ErrRecordNotFound {
			hwlog.RunLog.Errorf("get node group info, id %v do no exist", id)
			return common.RespMsg{Status: "",
				Msg: fmt.Sprintf("get node group info, id %v do no exist", id), Data: nil}
		}
		if err != nil {
			hwlog.RunLog.Errorf("get node group info id %v, db failed", id)
			return common.RespMsg{Status: "",
				Msg: fmt.Sprintf("get node group info id %v, db failed", id), Data: nil}
		}
		nodeGroupInfos = append(nodeGroupInfos, types.NodeGroupInfo{
			NodeGroupID:   nodeGroupInfo.ID,
			NodeGroupName: nodeGroupInfo.GroupName,
		})
	}
	resp := types.InnerGetNodeGroupInfosResp{
		NodeGroupInfos: nodeGroupInfos,
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

func getAppInstanceCountByGroupId(groupId uint64) (int64, error) {
	router := common.Router{
		Source:      common.NodeManagerName,
		Destination: common.AppManagerName,
		Option:      common.Get,
		Resource:    common.AppInstanceByNodeGroup,
	}
	resp := common.SendSyncMessageByRestful([]uint64{groupId}, &router)
	if resp.Status != common.Success {
		return 0, errors.New(resp.Msg)
	}
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return 0, errors.New("marshal internal response error")
	}
	counts := make(map[uint64]int64)
	if err = json.Unmarshal(data, &counts); err != nil {
		return 0, errors.New("unmarshal internal response error")
	}
	count, ok := counts[groupId]
	if !ok {
		return 0, errors.New("can't find corresponding groupId")
	}
	return count, nil
}
