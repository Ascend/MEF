// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"encoding/json"
	"github.com/gin-gonic/gin/binding"
)

// CreateEdgeNodeReq Create edge node
type CreateEdgeNodeReq struct {
	Description string `json:"description,omitempty"`
	NodeName    string `json:"nodeName"`
	UniqueName  string `json:"uniqueName"`
	NodeGroup   string `json:"nodeGroup,omitempty"`
}

// CreateNodeGroupReq Create edge node group
type CreateNodeGroupReq struct {
	Description   string `json:"description,omitempty"`
	NodeGroupName string `json:"nodeGroupName"`
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

// BatchDeleteNodeRelationReq delete node relation
type BatchDeleteNodeRelationReq struct {
	GroupID int64   `json:"groupId"`
	NodeIDs []int64 `json:"nodeId"`
}

// Check request validator
func (req BatchDeleteNodeRelationReq) Check() error {
	return nil
}

// GetNodeDetailResp nodeDetail
type GetNodeDetailResp struct {
	Id          int64  `json:"id"`
	NodeName    string `json:"nodeName"`
	UniqueName  string `json:"uniqueName"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	Cpu         int64  `json:"cpu"`
	Memory      int64  `json:"memory"`
	Npu         string `json:"npu"`
	NodeType    string `json:"nodeType"`
	NodeGroup   string `json:"nodeGroup"`
}

// ModifyNodeGroupReq request object
type ModifyNodeGroupReq struct {
	NodeId      int64  `json:"nodeId"`
	NodeName    string `json:"nodeName"`
	Description string `json:"description"`
}

// Check request validator
func (req ModifyNodeGroupReq) Check() error {
	return nil
}

// GetNodeStatisticsResp node statistics data
type GetNodeStatisticsResp = map[string]int64

// ListNodeGroupResp response object for listNodeGroup
type ListNodeGroupResp struct {
	Total  int64                   `json:"total"`
	Groups []ListNodeGroupRespItem `json:"groups"`
}

// ListNodeGroupRespItem group data
type ListNodeGroupRespItem struct {
	GroupID       int64  `json:"groupId"`
	NodeGroupName string `json:"nodeGroupName"`
	Description   string `json:"description"`
	CreateAt      string `json:"createAt"`
	NodeCount     int64  `json:"nodeCount"`
}

// GetNodeGroupDetailResp response object for nodeGroupDetail
type GetNodeGroupDetailResp struct {
	Nodes []GetNodeGroupDetailRespItem `json:"nodes"`
}

// GetNodeGroupDetailRespItem Node data
type GetNodeGroupDetailRespItem struct {
	NodeID      int64  `json:"nodeId"`
	NodeName    string `json:"nodeName"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreateAt    string `json:"createAt"`
	UpdateAt    string `json:"updateAt"`
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

// ListNodesResp list nodes response
type ListNodesResp struct {
	Nodes *[]NodeInfo `json:"nodes"`
	Total int         `json:"total"`
}

// BatchDeleteNodeGroupReq batch delete node group
type BatchDeleteNodeGroupReq struct {
	GroupID []int64 `json:"groupID"`
}
