// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package types defines structs which could be used in different package
package types

// InnerGetNodeInfoByNameReq is the request struct for internal module to get node info by node name
type InnerGetNodeInfoByNameReq struct {
	UniqueName string `json:"uniqueName"`
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
	NodeID        uint64            `json:"nodeID"`
	NodeName      string            `json:"nodeName"`
	UniqueName    string            `json:"uniqueName"`
	SoftwareInfos map[string]string `json:"softwareInfos"`
}

// InnerGetNodeGroupInfosResp is the response struct of node group infos by group ids
type InnerGetNodeGroupInfosResp struct {
	NodeGroupInfos []NodeGroupInfo `json:"nodeGroupInfos"`
}

// InnerGetNodeStatusResp is the response struct of node status
type InnerGetNodeStatusResp struct {
	NodeStatus string `json:"nodeStatus"`
}

// EdgeReportSoftwareInfoReq [struct] to report edge software info
type EdgeReportSoftwareInfoReq struct {
	UniqueName    string            `json:"uniqueName"`
	SoftwareInfos map[string]string `json:"softwareInfos"`
}
