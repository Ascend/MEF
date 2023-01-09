// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"huawei.com/mindxedge/base/common"
)

const (
	paramNodeName          = "nodeName"
	paramNodeNameShort     = "name"
	paramUniqueName        = "uniqueName"
	paramNodeGroupName     = "nodeGroup"
	paramNodeGroupNameLong = "nodeGroupName"
	paramDescription       = "description"
)

// CreateEdgeNodeReq Create edge node
type CreateEdgeNodeReq struct {
	Description string `json:"description,omitempty"`
	NodeName    string `json:"nodeName"`
	UniqueName  string `json:"uniqueName"`
	NodeGroup   string `json:"nodeGroup,omitempty"`
}

// Check request validator
func (req CreateEdgeNodeReq) Check() error {
	return common.NewValidator().
		ValidateNodeName(paramNodeName, req.NodeName).
		ValidateNodeUniqueName(paramUniqueName, req.UniqueName).
		Error()
}

// CreateNodeGroupReq Create edge node group
type CreateNodeGroupReq struct {
	Description   string `json:"description,omitempty"`
	NodeGroupName string `json:"nodeGroupName"`
}

// Check request validator
func (req CreateNodeGroupReq) Check() error {
	return common.NewValidator().
		ValidateNodeGroupName(paramNodeGroupNameLong, req.NodeGroupName).
		ValidateNodeGroupDesc(paramDescription, req.Description).
		Error()
}

// GetNodeDetailReq request object
type GetNodeDetailReq struct {
	Id int64 `json:"id" uri:"id"`
}

// GetNodeGroupDetailReq request object
type GetNodeGroupDetailReq = GetNodeDetailReq

// Check request validator
func (req GetNodeDetailReq) Check() error {
	return nil
}

// BatchDeleteNodeReq batch delete node
type BatchDeleteNodeReq []int64

// Check request validator
func (req BatchDeleteNodeReq) Check() error {
	return nil
}

// DeleteNodeToGroupReq delete nodes to group
type DeleteNodeToGroupReq struct {
	GroupID int64   `json:"groupID"`
	NodeIDs []int64 `json:"nodeIDs"`
}

// BatchDeleteNodeRelationReq delete multiple node-group relation
type BatchDeleteNodeRelationReq []DeleteNodeRelationReq

// DeleteNodeRelationReq delete single node-group relation
type DeleteNodeRelationReq struct {
	GroupID int64 `json:"groupID"`
	NodeID  int64 `json:"nodeID"`
}

// Check request validator
func (req BatchDeleteNodeRelationReq) Check() error {
	return nil
}

// ModifyNodeReq request object
type ModifyNodeReq struct {
	NodeID      int64  `json:"nodeID"`
	NodeName    string `json:"nodeName"`
	Description string `json:"description"`
}

// Check request validator
func (req ModifyNodeReq) Check() error {
	return common.NewValidator().
		ValidateNodeName(paramNodeName, req.NodeName).
		ValidateNodeDesc(paramDescription, req.Description).
		Error()
}

// ModifyNodeGroupReq request object
type ModifyNodeGroupReq struct {
	GroupID     int64  `json:"groupID"`
	GroupName   string `json:"nodeGroupName"`
	Description string `json:"description"`
}

// Check request validator
func (req ModifyNodeGroupReq) Check() error {
	return common.NewValidator().
		ValidateNodeGroupName(paramNodeGroupNameLong, req.GroupName).
		ValidateNodeGroupDesc(paramDescription, req.Description).
		Error()
}

// AddNodeToGroupReq Create edge node group
type AddNodeToGroupReq struct {
	NodeIDs []int64 `json:"nodeIDs"`
	GroupID int64   `json:"groupID"`
}

// AddUnManagedNodeReq add unmanaged node
type AddUnManagedNodeReq struct {
	NodeID      int64   `json:"nodeID"`
	NodeName    string  `json:"name"`
	GroupIDs    []int64 `json:"groupIDs,omitempty"`
	Description string  `json:"description,omitempty"`
}

// Check request validator
func (req AddUnManagedNodeReq) Check() error {
	return common.NewValidator().
		ValidateNodeName(paramNodeNameShort, req.NodeName).
		ValidateNodeDesc(paramDescription, req.Description).
		Error()
}

// BatchDeleteNodeGroupReq batch delete node group
type BatchDeleteNodeGroupReq struct {
	GroupIDs []int64 `json:"groupIDs"`
}

// Check request validator
func (req BatchDeleteNodeGroupReq) Check() error {
	return nil
}

// ListNodeGroupResp response object for listNodeGroup
type ListNodeGroupResp struct {
	Total  int64         `json:"total"`
	Groups []NodeGroupEx `json:"groups"`
}

// ListNodesResp list managed nodes response
type ListNodesResp struct {
	Nodes *[]NodeInfoDetail `json:"nodes"`
	Total int               `json:"total"`
}

// ListNodesUnmanagedResp list unmanaged nodes response
type ListNodesUnmanagedResp struct {
	Nodes *[]NodeInfoEx `json:"nodes"`
	Total int           `json:"total"`
}

// NodeGroupDetail get node group detail response
type NodeGroupDetail struct {
	NodeGroup
	Nodes []NodeInfoEx `json:"nodes"`
}

// NodeGroupEx contains node group and nodes count
type NodeGroupEx struct {
	NodeGroup
	NodeCount int64 `json:"nodeCount"`
}

// Extend construct NodeGroupEx
func (n *NodeGroupEx) Extend(nodeGroup *NodeGroup, nodeCount int64) {
	*n = NodeGroupEx{
		NodeGroup: *nodeGroup,
		NodeCount: nodeCount,
	}
}

// NodeInfoEx contains static info and dynamic info
type NodeInfoEx struct {
	NodeInfo
	NodeInfoDynamic
}

// Extend construct NodeInfoEx
func (n *NodeInfoEx) Extend(info *NodeInfo, dynamicInfo *NodeInfoDynamic) {
	if dynamicInfo == nil {
		dynamicInfo = &NodeInfoDynamic{Status: statusOffline}
	}
	*n = NodeInfoEx{
		NodeInfo:        *info,
		NodeInfoDynamic: *dynamicInfo,
	}
}

// NodeInfoDetail contains static info, dynamic info and group names
type NodeInfoDetail struct {
	NodeInfoEx
	NodeGroup string `json:"nodeGroup"`
}

// Extend construct NodeInfoDetail
func (n *NodeInfoDetail) Extend(info *NodeInfo, dynamicInfo *NodeInfoDynamic, nodeGroup string) {
	var nodeInfoEx NodeInfoEx
	nodeInfoEx.Extend(info, dynamicInfo)
	*n = NodeInfoDetail{
		NodeInfoEx: nodeInfoEx,
		NodeGroup:  nodeGroup,
	}
}
