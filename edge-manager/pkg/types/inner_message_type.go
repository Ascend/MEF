// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package types defines structs which could be used in different package
package types

// InnerGetNodeInfoByNameReq is the request struct for internal module to get node info by node name
type InnerGetNodeInfoByNameReq struct {
	UniqueName string `json:"uniqueName"`
}

// InnerGetNodeGroupInfoByIdReq is the request struct for internal module to get node group info by group id
type InnerGetNodeGroupInfoByIdReq struct {
	GroupID int64 `json:"groupId"`
}

// InnerGetNodeStatusReq is request struct for internal module to get node status
type InnerGetNodeStatusReq struct {
	UniqueName string `json:"uniqueName"`
}

// InnerGetNodeInfoByNameResp is the response struct of node info
type InnerGetNodeInfoByNameResp struct {
	NodeID int64 `json:"nodeId"`
}

// InnerGetNodeGroupInfoByIdResp is the response struct of node group info
type InnerGetNodeGroupInfoByIdResp struct {
	GroupName string `json:"groupName"`
}

// InnerGetNodeStatusResp is the response struct of node status
type InnerGetNodeStatusResp struct {
	NodeStatus string `json:"nodeStatus"`
}
