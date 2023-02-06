// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

// CreateEdgeNodeReq Create edge node
type CreateEdgeNodeReq struct {
	Description string   `json:"description,omitempty"`
	NodeName    *string  `json:"nodeName"`
	UniqueName  *string  `json:"uniqueName"`
	GroupIDs    []uint64 `json:"nodeGroup,omitempty"`
}

// CreateNodeGroupReq Create edge node group
type CreateNodeGroupReq struct {
	Description   string  `json:"description,omitempty"`
	NodeGroupName *string `json:"nodeGroupName"`
}

// BatchDeleteNodeReq batch delete node
type BatchDeleteNodeReq struct {
	NodeIDs []uint64 `json:"nodeIDs"`
}

// DeleteNodeToGroupReq delete nodes to group
type DeleteNodeToGroupReq struct {
	GroupID *uint64   `json:"groupID"`
	NodeIDs *[]uint64 `json:"nodeIDs"`
}

// BatchDeleteNodeRelationReq delete multiple node-group relation
type BatchDeleteNodeRelationReq []DeleteNodeRelationReq

// DeleteNodeRelationReq delete single node-group relation
type DeleteNodeRelationReq struct {
	GroupID *uint64 `json:"groupID"`
	NodeID  *uint64 `json:"nodeID"`
}

// ModifyNodeReq request object
type ModifyNodeReq struct {
	NodeID      *uint64 `json:"nodeID"`
	NodeName    *string `json:"nodeName"`
	Description string  `json:"description"`
}

// ModifyNodeGroupReq request object
type ModifyNodeGroupReq struct {
	GroupID     *uint64 `json:"groupID"`
	GroupName   *string `json:"nodeGroupName"`
	Description string  `json:"description"`
}

// AddNodeToGroupReq Create edge node group
type AddNodeToGroupReq struct {
	NodeIDs *[]uint64 `json:"nodeIDs"`
	GroupID *uint64   `json:"groupID"`
}

// AddUnManagedNodeReq add unmanaged node
type AddUnManagedNodeReq struct {
	NodeID      *uint64  `json:"nodeID"`
	NodeName    *string  `json:"name"`
	GroupIDs    []uint64 `json:"groupIDs,omitempty"`
	Description string   `json:"description,omitempty"`
}

// BatchDeleteNodeGroupReq batch delete node group
type BatchDeleteNodeGroupReq struct {
	GroupIDs *[]uint64 `json:"groupIDs"`
}

// ListNodeGroupResp response object for listNodeGroup
type ListNodeGroupResp struct {
	Total  int64         `json:"total"`
	Groups []NodeGroupEx `json:"groups"`
}

// ListNodesResp list managed nodes response
type ListNodesResp struct {
	Nodes []NodeInfoExManaged `json:"nodes"`
	Total int                 `json:"total"`
}

// ListNodesUnmanagedResp list unmanaged nodes response
type ListNodesUnmanagedResp struct {
	Nodes []NodeInfoEx `json:"nodes"`
	Total int          `json:"total"`
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

// NodeInfoEx node information for unmanaged node
type NodeInfoEx struct {
	NodeInfo
	Status string `json:"status"`
}

// NodeInfoExManaged node information for managed node
type NodeInfoExManaged struct {
	NodeInfoEx
	NodeGroup string `json:"nodeGroup"`
}

// NodeInfoDetail contains static info, dynamic info and group names
type NodeInfoDetail struct {
	NodeInfoExManaged
	NodeResource
	Npu int64 `json:"npu"`
}
