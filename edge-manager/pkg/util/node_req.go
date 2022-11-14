// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package util to init util service
package util

// CreateEdgeNodeReq Create edge node
type CreateEdgeNodeReq struct {
	Description string `json:"description"`
	NodeName    string `json:"nodeName"`
	UniqueName  string `json:"uniqueName"`
	NodeGroup   string `json:"nodeGroup,omitempty"`
}

// CreateNodeGroupReq Create edge node group
type CreateNodeGroupReq struct {
	Description   string `json:"description"`
	NodeGroupName string `json:"nodeGroupName"`
}
