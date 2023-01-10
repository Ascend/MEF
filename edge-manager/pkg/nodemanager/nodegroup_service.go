// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node database table
package nodemanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/types"
	"edge-manager/pkg/util"
)

// CreateGroup Create Node Group
func createGroup(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start create node group")
	var req CreateNodeGroupReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("create node group convert request error, %s", err.Error())
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	if checkResult := newCreateGroupChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("create node group validate parameters error, %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}

	checker := specificationChecker{nodeService: NodeServiceInstance()}
	if err := checker.checkAddGroups(1); err != nil {
		hwlog.RunLog.Errorf("create node group check spec error: %s", err.Error())
		return common.RespMsg{Msg: err.Error()}
	}
	group := &NodeGroup{
		Description: req.Description,
		GroupName:   *req.NodeGroupName,
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdatedAt:   time.Now().Format(TimeFormat),
	}
	if err := NodeServiceInstance().createNodeGroup(group); err != nil {
		if strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
			hwlog.RunLog.Error("node group is duplicate")
			return common.RespMsg{Status: "", Msg: "node group is duplicate", Data: nil}
		}
		hwlog.RunLog.Error("node group db create failed")
		return common.RespMsg{Status: "", Msg: "db group create failed", Data: nil}
	}
	hwlog.RunLog.Info("node group db create success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func listEdgeNodeGroup(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start list node group")
	req, ok := input.(util.ListReq)
	if !ok {
		return common.RespMsg{Status: "", Msg: "convert request error", Data: nil}
	}
	var resp ListNodeGroupResp
	count, err := NodeServiceInstance().countNodeGroupsByName(req.Name)
	if err != nil {
		hwlog.RunLog.Error("node group db query failed")
		return common.RespMsg{Status: "", Msg: "db group query failed", Data: nil}
	}
	resp.Total = count
	nodeGroups, err := NodeServiceInstance().getNodeGroupsByName(req.PageNum, req.PageSize, req.Name)
	if err != nil {
		hwlog.RunLog.Error("node group db query failed")
		return common.RespMsg{Status: "", Msg: "db group query failed", Data: nil}
	}
	for _, group := range *nodeGroups {
		relations, err := NodeServiceInstance().listNodeRelationsByGroupId(group.ID)
		if err != nil {
			hwlog.RunLog.Error("node group db query failed")
			return common.RespMsg{Status: "", Msg: "db group query failed", Data: nil}
		}
		respItem := NodeGroupEx{
			NodeGroup: group,
			NodeCount: int64(len(*relations)),
		}
		resp.Groups = append(resp.Groups, respItem)
	}
	hwlog.RunLog.Info("node group db query success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

func getEdgeNodeGroupDetail(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start get node group detail")
	var req GetNodeGroupDetailReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("get node group detail convert request error, %s", err.Error())
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	if checkResult := newGetGroupDetailChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("get node detail check parameters failed, %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}
	var resp NodeGroupDetail
	nodeGroup, err := NodeServiceInstance().getNodeGroupByID(req.ID)
	if err != nil {
		hwlog.RunLog.Error("node group db query failed")
		return common.RespMsg{Status: "", Msg: "db query failed", Data: nil}
	}
	resp.NodeGroup = *nodeGroup
	relations, err := NodeServiceInstance().listNodeRelationsByGroupId(req.ID)
	if err != nil {
		hwlog.RunLog.Error("node group db query failed")
		return common.RespMsg{Status: "", Msg: "db query failed", Data: nil}
	}
	for _, relation := range *relations {
		node, err := NodeServiceInstance().getNodeByID(relation.NodeID)
		if err != nil {
			hwlog.RunLog.Error("node group db query failed")
			return common.RespMsg{Status: "", Msg: "db query failed", Data: nil}
		}
		nodeInfoDynamic, _ := NodeStatusServiceInstance().Get(node.UniqueName)
		var nodeInfoEx NodeInfoEx
		nodeInfoEx.Extend(node, nodeInfoDynamic)
		resp.Nodes = append(resp.Nodes, nodeInfoEx)
	}
	hwlog.RunLog.Info("node group db query success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

func modifyNodeGroup(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start modify node group")
	var req ModifyNodeGroupReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("modify node group convert request error, %s", err.Error())
		return common.RespMsg{Status: "", Msg: "convert request error", Data: nil}
	}
	if checkResult := newModifyGroupChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("modify node group check parameters failed, %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}
	updatedColumns := map[string]interface{}{
		"GroupName":   req.GroupName,
		"Description": req.Description,
		"UpdatedAt":   time.Now().Format(TimeFormat),
	}
	if count, err := NodeServiceInstance().updateGroup(*req.GroupID, updatedColumns); err != nil || count != 1 {
		if err != nil && strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
			hwlog.RunLog.Error("node group name is duplicate")
			return common.RespMsg{Status: "", Msg: "node group name is duplicate", Data: nil}
		}
		hwlog.RunLog.Error("modify node group db update error")
		return common.RespMsg{Status: "", Msg: "db update error", Data: nil}
	}
	hwlog.RunLog.Info("modify node group db update success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func batchDeleteNodeGroup(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start batch delete node group")
	var req BatchDeleteNodeGroupReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("batch delete node group convert request error, %s", err)
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	if checkResult := newBatchDeleteGroupChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("batch delete node group check parameters failed, %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}
	var delSuccessGroupID []int64
	for _, groupID := range *req.GroupIDs {
		if err := deleteSingleGroup(groupID); err != nil {
			hwlog.RunLog.Errorf("delete node group %d failed, %s", groupID, err.Error())
			continue
		}
		delSuccessGroupID = append(delSuccessGroupID, groupID)
	}
	if len(*req.GroupIDs) > len(delSuccessGroupID) {
		hwlog.RunLog.Error("batch delete node group failed")
		return common.RespMsg{Status: "", Msg: "batch delete node group failed", Data: delSuccessGroupID}
	}
	hwlog.RunLog.Info("batch delete node group success")
	return common.RespMsg{Status: common.Success, Msg: "batch delete node group success", Data: delSuccessGroupID}
}

func deleteSingleGroup(groupID int64) error {
	nodeGroup, err := NodeServiceInstance().getNodeGroupByID(groupID)
	if err != nil {
		return fmt.Errorf("get node group by group id %d failed", groupID)
	}
	count, err := getAppInstanceCountByGroupId(groupID)
	if err != nil {
		return err
	}
	if count != 0 {
		return fmt.Errorf("group %d has deployed app, can't remove", groupID)
	}
	relations, err := NodeServiceInstance().listNodeRelationsByGroupId(groupID)
	if err != nil {
		return fmt.Errorf("get relations between node and node group by group id %d failed", groupID)
	}
	for _, relation := range *relations {
		if err := deleteSingleNodeRelation(nodeGroup.ID, relation.NodeID); err != nil {
			return fmt.Errorf("patch node state failed:%s", err.Error())
		}
	}
	if rowsAffected, err := NodeServiceInstance().deleteNodeGroup(groupID); err != nil || rowsAffected != 1 {
		return fmt.Errorf("delete node group by group id %d failed", groupID)
	}
	return nil
}

func getAppInstanceCountByGroupId(groupId int64) (int64, error) {
	router := common.Router{
		Source:      common.NodeManagerName,
		Destination: common.AppManagerName,
		Option:      common.Get,
		Resource:    common.AppInstanceByNodeGroup,
	}
	resp := common.SendSyncMessageByRestful([]int64{groupId}, &router)
	if resp.Status != common.Success {
		return 0, errors.New(resp.Msg)
	}
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return 0, errors.New("convert result failed")
	}
	counts := make(map[int64]int64)
	if err = json.Unmarshal(data, &counts); err != nil {
		return 0, errors.New("convert result failed")
	}
	count, ok := counts[groupId]
	if !ok {
		return 0, errors.New("can't find corresponding groupId")
	}
	return count, nil
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
