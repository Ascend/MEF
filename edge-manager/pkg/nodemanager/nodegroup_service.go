// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node database table
package nodemanager

import (
	"strings"
	"time"

	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/util"
	"huawei.com/mindxedge/base/common"
)

// CreateGroup Create Node Group
func createGroup(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start create node group")
	var req CreateNodeGroupReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Error("create node group conver request error")
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	if err := req.Check(); err != nil {
		hwlog.RunLog.Error("create node group validate parameters error")
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: err.Error()}
	}
	total, err := GetTableCount(NodeGroup{})
	if err != nil {
		hwlog.RunLog.Error("get node group table num failed")
		return common.RespMsg{Status: "", Msg: "get node group table num failed", Data: nil}
	}
	if total >= maxNodeGroup {
		hwlog.RunLog.Error("node group number is enough, connot create")
		return common.RespMsg{Status: "", Msg: "node group number is enough, connot create", Data: nil}
	}
	group := &NodeGroup{
		Description: req.Description,
		GroupName:   req.NodeGroupName,
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdateAt:    time.Now().Format(TimeFormat),
	}
	if err = NodeServiceInstance().createNodeGroup(group); err != nil {
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
		respItem := ListNodeGroupRespItem{
			GroupID:       group.ID,
			NodeGroupName: group.GroupName,
			Description:   group.Description,
			CreateAt:      group.CreatedAt,
			NodeCount:     int64(len(*relations)),
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
		hwlog.RunLog.Error("get node group detail convert request error")
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	var resp GetNodeGroupDetailResp
	relations, err := NodeServiceInstance().listNodeRelationsByGroupId(req.Id)
	if err != nil {
		hwlog.RunLog.Error("node group db query failed")
		return common.RespMsg{Status: "", Msg: "db query failed", Data: nil}
	}
	for _, relation := range *relations {
		node, err := NodeServiceInstance().GetNodeByID(relation.NodeID)
		if err != nil {
			hwlog.RunLog.Error("node group db query failed")
			return common.RespMsg{Status: "", Msg: "db query failed", Data: nil}
		}
		status := NodeStatusServiceInstance().GetNodeStatus(node.UniqueName)
		nodeResp := GetNodeGroupDetailRespItem{
			NodeID:      node.ID,
			NodeName:    node.NodeName,
			Description: node.Description,
			Status:      status,
			CreateAt:    node.CreatedAt,
			UpdateAt:    node.UpdateAt,
		}
		resp.Nodes = append(resp.Nodes, nodeResp)
	}
	hwlog.RunLog.Info("node group db query success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: resp}
}
