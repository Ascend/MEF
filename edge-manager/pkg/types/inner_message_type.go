// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package types defines structs which could be used in different package
package types

// InnerGetNodesInfoByNameReq is the request struct for internal module to get node info by node name
type InnerGetNodesInfoByNameReq struct {
	UniqueNames []string `json:"uniqueName"`
}

// InnerGetNodeGroupInfosReq is the request struct for internal module to get node group infos by group ids
type InnerGetNodeGroupInfosReq struct {
	NodeGroupIds []uint64 `json:"nodeGroupIDs"`
}

// InnerGetNodeStatusReq is request struct for internal module to get node status
type InnerGetNodeStatusReq struct {
	UniqueName string `json:"uniqueName"`
}

// InnerGetNodeInfoByNameResp is the response struct of node info
type InnerGetNodeInfoByNameResp struct {
	NodeID       uint64            `json:"nodeID"`
	NodeName     string            `json:"nodeName"`
	UniqueName   string            `json:"uniqueName"`
	VersionInfos map[string]string `json:"versionInfos"`
}

// InnerGetNodeGroupInfosResp is the response struct of node group infos by group ids
type InnerGetNodeGroupInfosResp struct {
	NodeGroupInfos []NodeGroupInfo `json:"nodeGroupInfos"`
}

// InnerGetNodeStatusResp is the response struct of node status
type InnerGetNodeStatusResp struct {
	NodeStatus string `json:"nodeStatus"`
}
