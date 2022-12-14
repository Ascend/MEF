// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"encoding/json"

	"github.com/gin-gonic/gin/binding"

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
		ValidateNodeGroupName(paramNodeGroupName, req.NodeGroup).
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

// UnmarshalJSON custom JSON unmarshal
func (req *GetNodeDetailReq) UnmarshalJSON(input []byte) error {
	return unmarshalUriParams(input, req)
}

func unmarshalUriParams(input []byte, obj interface{}) error {
	objMap := make(map[string][]string)
	if err := json.Unmarshal(input, &objMap); err != nil {
		return err
	}
	if err := binding.Uri.BindUri(objMap, obj); err != nil {
		return err
	}
	return nil
}

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
	GroupID int64   `json:"groupId"`
	NodeIDs []int64 `json:"nodeId"`
}

// BatchDeleteNodeRelationReq delete multiple node-group relation
type BatchDeleteNodeRelationReq []DeleteNodeRelationReq

// DeleteNodeRelationReq delete single node-group relation
type DeleteNodeRelationReq struct {
	GroupID int64 `json:"groupId"`
	NodeID  int64 `json:"nodeId"`
}

// Check request validator
func (req BatchDeleteNodeRelationReq) Check() error {
	return nil
}

// ModifyNodeReq request object
type ModifyNodeReq struct {
	NodeId      int64  `json:"nodeId"`
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
	GroupId     int64  `json:"groupId"`
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
	NodeID  []int64 `json:"nodeId"`
	GroupID int64   `json:"groupId"`
}

// AddUnManagedNodeReq add unmanaged node
type AddUnManagedNodeReq struct {
	NodeID      int64   `json:"nodeId"`
	NodeName    string  `json:"name"`
	GroupID     []int64 `json:"groupId,omitempty"`
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
	GroupID []int64 `json:"groupID"`
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
