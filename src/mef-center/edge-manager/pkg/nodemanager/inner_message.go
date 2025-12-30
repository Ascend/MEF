// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package nodemanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"k8s.io/api/core/v1"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-manager/pkg/types"

	"huawei.com/mindxedge/base/common/requests"

	"huawei.com/mindxedge/base/common"
)

func innerGetNodeInfoByUniqueName(msg *model.Message) common.RespMsg {
	var req types.InnerGetNodeInfoByNameReq
	if err := msg.ParseContent(&req); err != nil {
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

func innerGetNodeSoftwareInfo(msg *model.Message) common.RespMsg {
	var req types.InnerGetSfwInfoBySNReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Error("parse inner message content failed")
		return common.RespMsg{Status: "", Msg: "parse inner message content failed", Data: nil}
	}
	nodeInfo, err := NodeServiceInstance().getNodeInfoBySerialNumber(req.SerialNumber)
	if err != nil {
		hwlog.RunLog.Errorf("get node info by unique name [%s] failed: %v", req.SerialNumber, err)
		return common.RespMsg{Status: "", Msg: "get node info by unique name failed", Data: nil}
	}
	resp := types.InnerSoftwareInfoResp{}

	if err = json.Unmarshal([]byte(nodeInfo.SoftwareInfo), &resp.SoftwareInfo); err != nil {
		hwlog.RunLog.Errorf("unmarshal node software info failed: %v", err)
		return common.RespMsg{Status: "", Msg: "get node info failed because unmarshal failed", Data: nil}
	}

	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

func innerGetNodeStatus(msg *model.Message) common.RespMsg {
	var req types.InnerGetNodeStatusReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Error("parse inner message content failed")
		return common.RespMsg{Status: "", Msg: "parse inner message content failed", Data: nil}
	}
	status, err := NodeSyncInstance().GetK8sNodeStatus(req.UniqueName)
	if err != nil {
		hwlog.RunLog.Error("inner message get node status failed")
		return common.RespMsg{Status: "", Msg: "internal get node status failed", Data: nil}
	}
	resp := types.InnerGetNodeStatusResp{
		NodeStatus: status,
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

func innerGetNodesByNodeGroupID(msg *model.Message) common.RespMsg {
	var req types.InnerGetNodesReq
	if err := msg.ParseContent(&req); err != nil {
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

func innerGetNodeSnsByGroupId(msg *model.Message) common.RespMsg {
	var reqInfo requests.GetSnsReq
	var inputInfo string
	if err := msg.ParseContent(&inputInfo); err != nil {
		hwlog.RunLog.Error("failed to convert param into string")
		return common.RespMsg{Status: common.ErrorParamConvert}
	}
	decoder := json.NewDecoder(strings.NewReader(inputInfo))
	decoder.UseNumber()
	err := decoder.Decode(&reqInfo)
	if err != nil {
		hwlog.RunLog.Error("failed to decode param into string")
		return common.RespMsg{Status: common.ErrorParamConvert}
	}

	gpId := reqInfo.GroupId
	_, err = NodeServiceInstance().getNodeGroupByID(gpId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hwlog.RunLog.Error("node group with specific id not found")
		return common.RespMsg{Status: common.ErrorNodeGroupNotFound, Data: nil}
	}
	if err != nil {
		hwlog.RunLog.Errorf("failed to get node group,err:%v", err)
		return common.RespMsg{Status: common.ErrorGetNodeGroup}
	}
	relations, err := NodeServiceInstance().listNodeRelationsByGroupId(gpId)
	if err != nil {
		hwlog.RunLog.Error("node group db query failed")
		return common.RespMsg{Status: common.ErrorGetNodeGroup, Msg: "list nodes by group in db failed", Data: nil}
	}
	nodeSns := make([]string, 0)
	for _, relation := range *relations {
		node, err := NodeServiceInstance().getNodeByID(relation.NodeID)
		if err != nil {
			hwlog.RunLog.Errorf("query node group db by id(%d) failed", relation.NodeID)
			return common.RespMsg{Status: common.ErrorGetNodeGroup, Msg: "query node by relations failed", Data: nil}
		}
		nodeSns = append(nodeSns, node.SerialNumber)
	}
	hwlog.RunLog.Info("node group Sns query success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nodeSns}
}

func innerGetIpBySn(msg *model.Message) common.RespMsg {
	var sn string
	if err := msg.ParseContent(&sn); err != nil {
		hwlog.RunLog.Error("parse inner message content failed")
		return common.RespMsg{Status: "", Msg: "parse inner message content failed"}
	}

	node, err := NodeServiceInstance().getNodeBySn(sn)
	if err != nil {
		hwlog.RunLog.Errorf("inner message get node info by sn failed:%s", err.Error())
		return common.RespMsg{Status: "", Msg: "inner message get node info by sn failed"}
	}

	return common.RespMsg{Status: common.Success, Msg: "", Data: node.IP}
}

func innerAllNodeInfos(*model.Message) common.RespMsg {
	nodeInfos, err := NodeServiceInstance().listNodes()
	if err != nil {
		hwlog.RunLog.Error("inner message get all node info failed")
		return common.RespMsg{Status: "", Msg: "internal get all node info failed", Data: nil}
	}
	hwlog.RunLog.Info("inner message get all node info success")
	return common.RespMsg{Status: common.Success, Msg: "internal get all node info success", Data: nodeInfos}
}

func innerCheckNodeGroupResReq(msg *model.Message) common.RespMsg {
	var req types.InnerCheckNodeResReq
	if err := msg.ParseContent(&req); err != nil {
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
		if err = checkNodeResource(req.ResourceReqs, nodeRelation.NodeID); err != nil {
			return fmt.Errorf("in group [%d], %s", req.NodeGroupID, err.Error())
		}
		if err = checkNodePodLimit(1, nodeRelation.NodeID); err != nil {
			return fmt.Errorf("in group [%d], %s", req.NodeGroupID, err.Error())
		}
	}
	return nil
}

func innerUpdateNodeGroupResReq(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start to update node group allocated resources")
	var req types.InnerUpdateNodeResReq
	if err := msg.ParseContent(&req); err != nil {
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
	var count int
	for name, quantity := range req {
		if count > common.MaxLoopNum {
			break
		}
		count++
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

func innerGetNodeGroupInfosByIds(msg *model.Message) common.RespMsg {
	var req types.InnerGetNodeGroupInfosReq
	if err := msg.ParseContent(&req); err != nil {
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
	resp := common.SendSyncMessageByRestful([]uint64{groupId}, &router, common.ResponseTimeout)
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

func innerGetNodeSnAndIpByID(msg *model.Message) common.RespMsg {
	var req types.InnerGetNodeInfosReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Error("parse inner message content failed")
		return common.RespMsg{Status: "", Msg: "parse inner message content failed", Data: nil}
	}
	var nodeInfos []types.NodeInfo
	for _, id := range req.NodeIds {
		nodeInfo, err := NodeServiceInstance().getNodeByID(id)
		if err == gorm.ErrRecordNotFound {
			hwlog.RunLog.Errorf("get node info, id %v do no exist", id)
			continue
		}
		if err != nil {
			hwlog.RunLog.Errorf("get node id %v, db failed", id)
			continue
		}
		nodeInfos = append(nodeInfos, types.NodeInfo{
			NodeID:       nodeInfo.ID,
			UniqueName:   nodeInfo.UniqueName,
			SerialNumber: nodeInfo.SerialNumber,
			Ip:           nodeInfo.IP,
		})
	}
	resp := types.InnerGetNodeInfosResp{
		NodeInfos: nodeInfos,
	}
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}
