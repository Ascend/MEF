// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node database table
package nodemanager

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/types"
	"edge-manager/pkg/util"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/logmgmt"
)

func createNodeGroup(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start create node group")
	var req CreateNodeGroupReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("create node group convert request error, %s", err.Error())
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}
	if checkResult := newCreateGroupChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("create node group validate parameters error, %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}

	checker := specificationChecker{nodeService: NodeServiceInstance()}
	if err := checker.checkAddGroups(1); err != nil {
		hwlog.RunLog.Errorf("create node group check spec error: %s", err.Error())
		return common.RespMsg{Status: common.ErrorCheckNodeMrgSize, Msg: err.Error()}
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
			return common.RespMsg{Status: common.ErrorNodeMrgDuplicate, Msg: "node group is duplicate", Data: nil}
		}
		hwlog.RunLog.Error("node group db create failed")
		return common.RespMsg{Status: common.ErrorCreateNodeGroup, Msg: "", Data: nil}
	}
	hwlog.RunLog.Info("node group db create success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: group.ID}
}

func listNodeGroup(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start list node group")
	req, ok := input.(types.ListReq)
	if !ok {
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "convert list request error"}
	}
	if checkResult := util.NewPaginationQueryChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("list node group para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}
	var resp ListNodeGroupResp
	count, err := NodeServiceInstance().countNodeGroupsByName(req.Name)
	if err != nil {
		hwlog.RunLog.Error("node group db query failed")
		return common.RespMsg{Status: common.ErrorListNodeGroups, Msg: "count node groups in db failed"}
	}
	resp.Total = count
	nodeGroups, err := NodeServiceInstance().getNodeGroupsByName(req.PageNum, req.PageSize, req.Name)
	if err != nil {
		hwlog.RunLog.Error("node group db query failed")
		return common.RespMsg{Status: common.ErrorListNodeGroups, Msg: "query node group in db failed"}
	}
	for _, group := range *nodeGroups {
		relations, err := NodeServiceInstance().listNodeRelationsByGroupId(group.ID)
		if err != nil {
			hwlog.RunLog.Error("list nodes by group in db failed")
			return common.RespMsg{Status: common.ErrorListNodeGroups, Msg: "list nodes by group in db failed"}
		}
		respItem := NodeGroupEx{
			NodeGroup: group,
			NodeCount: int64(len(*relations)),
		}
		resp.Groups = append(resp.Groups, respItem)
	}
	hwlog.RunLog.Info("list node groups success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

func getNodeGroupDetail(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start get node group detail")
	id, ok := input.(uint64)
	if !ok {
		hwlog.RunLog.Error("get node group detail convert request error")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "convert node detail input failed"}
	}
	if checkResult := idChecker("").Check(id); !checkResult.Result {
		hwlog.RunLog.Errorf("get node group detail para check failed: %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason, Data: nil}
	}
	var resp NodeGroupDetail
	nodeGroup, err := NodeServiceInstance().getNodeGroupByID(id)
	if err != nil {
		hwlog.RunLog.Error("node group db query failed")
		return common.RespMsg{Status: common.ErrorGetNodeGroup, Msg: "nodegroup db query failed", Data: nil}
	}
	resp.NodeGroup = *nodeGroup
	relations, err := NodeServiceInstance().listNodeRelationsByGroupId(id)
	if err != nil {
		hwlog.RunLog.Error("node group db query failed")
		return common.RespMsg{Status: common.ErrorGetNodeGroup, Msg: "list nodes by group in db failed", Data: nil}
	}
	for _, relation := range *relations {
		var respItem NodeInfoEx
		node, err := NodeServiceInstance().getNodeByID(relation.NodeID)
		if err != nil {
			hwlog.RunLog.Errorf("query node group db by id(%d) failed", relation.NodeID)
			return common.RespMsg{Status: common.ErrorGetNodeGroup, Msg: "query node by relations failed", Data: nil}
		}
		respItem.NodeInfo = *node
		respItem.Status, err = NodeSyncInstance().GetMEFNodeStatus(node.UniqueName)
		if err != nil {
			respItem.Status = statusOffline
		}
		resp.Nodes = append(resp.Nodes, respItem)
	}
	hwlog.RunLog.Info("node group db query success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}

func modifyNodeGroup(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start modify node group")
	var req ModifyNodeGroupReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("modify node group convert request error, %s", err.Error())
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "convert modify node group request error"}
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
			return common.RespMsg{Status: common.ErrorNodeMrgDuplicate, Msg: "node group name is duplicate", Data: nil}
		}
		hwlog.RunLog.Error("modify node group db update error")
		return common.RespMsg{Status: common.ErrorModifyNodeGroup, Msg: "db update node group error", Data: nil}
	}
	hwlog.RunLog.Info("modify node group db update success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func getNodeGroupStatistics(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start get node group statistics")
	total, err := GetTableCount(NodeGroup{})
	if err != nil {
		hwlog.RunLog.Error("failed to get node group statistics, db query failed")
		return common.RespMsg{Status: common.ErrorCountNodeGroup, Msg: ""}
	}
	hwlog.RunLog.Info("get node group statistics success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: total}
}

func batchDeleteNodeGroup(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start batch delete node group")
	var req BatchDeleteNodeGroupReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Errorf("batch delete node group convert request error, %s", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}
	if checkResult := newBatchDeleteGroupChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("batch delete node group check parameters failed, %s", checkResult.Reason)
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkResult.Reason}
	}
	var delRes types.BatchResp
	failedMap := make(map[string]string)
	delRes.FailedInfos = failedMap
	for _, groupID := range *req.GroupIDs {
		if err := deleteSingleGroup(groupID); err != nil {
			hwlog.RunLog.Errorf("delete node group %d failed, %s", groupID, err.Error())
			failedMap[strconv.Itoa(int(groupID))] = fmt.Sprintf("delete failed, %s", err.Error())
			continue
		}
		delRes.SuccessIDs = append(delRes.SuccessIDs, groupID)
	}
	logmgmt.BatchOperationLog("batch delete node group", delRes.SuccessIDs)
	if len(delRes.FailedInfos) != 0 {
		hwlog.RunLog.Error("batch delete node group failed")
		return common.RespMsg{Status: common.ErrorDeleteNodeGroup, Msg: "", Data: delRes}
	}
	hwlog.RunLog.Info("batch delete node group success")
	return common.RespMsg{Status: common.Success, Msg: "batch delete node group success", Data: nil}
}

func deleteSingleGroup(groupID uint64) error {
	if _, err := NodeServiceInstance().getNodeGroupByID(groupID); err != nil {
		return fmt.Errorf("get node group by group id %d failed", groupID)
	}
	count, err := getAppInstanceCountByGroupId(groupID)
	if err != nil {
		hwlog.RunLog.Errorf("get from app error: %v", err)
		return err
	}
	if count != 0 {
		return fmt.Errorf("group %d has deployed app, can't remove", groupID)
	}
	relations, err := NodeServiceInstance().listNodeRelationsByGroupId(groupID)
	if err != nil {
		return fmt.Errorf("get relations between node and node group by group id %d failed", groupID)
	}
	if err = NodeServiceInstance().deleteNodeGroup(groupID, relations); err != nil {
		return err
	}
	return nil
}
