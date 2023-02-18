// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package types defines structs which could be used in different package
package types

import "k8s.io/api/core/v1"

// InnerGetNodeInfoByNameReq is the request struct for internal module to get node info by node name
type InnerGetNodeInfoByNameReq struct {
	UniqueName string `json:"uniqueName"`
}

// InnerGetSoftwareInfoBySerialNumberReq is the request struct for internal module to get node info by SerialNumber
type InnerGetSoftwareInfoBySerialNumberReq struct {
	SerialNumber string `json:"serialNumber"`
}

// InnerGetNodeGroupInfosReq is the request struct for internal module to get node group infos by group ids
type InnerGetNodeGroupInfosReq struct {
	NodeGroupIds []uint64 `json:"nodeGroupIDs"`
}

// InnerGetNodeStatusReq is request struct for internal module to get node status
type InnerGetNodeStatusReq struct {
	UniqueName string `json:"uniqueName"`
}

// InnerUpdateNodeResReq [struct] for CheckAndUpdateNodeResReq
type InnerUpdateNodeResReq struct {
	NodeGroupID  uint64 `json:"nodeGroupID"`
	ResourceReqs v1.ResourceList
	IsUndeploy   bool
}

// InnerCheckNodeResReq [struct] for UpdateNodeResReq
type InnerCheckNodeResReq struct {
	NodeGroupID  uint64 `json:"nodeGroupID"`
	ResourceReqs v1.ResourceList
}

// InnerSoftwareInfoResp is the response struct of node info
type InnerSoftwareInfoResp struct {
	SoftwareInfo []SoftwareInfo `json:"SoftwareInfo"`
}

// InnerGetNodeInfoByNameResp is the response struct of node info
type InnerGetNodeInfoByNameResp struct {
	NodeID   uint64 `json:"nodeID"`
	NodeName string `json:"nodeName"`
}

// InnerGetNodeGroupInfosResp is the response struct of node group infos by group ids
type InnerGetNodeGroupInfosResp struct {
	NodeGroupInfos []NodeGroupInfo `json:"nodeGroupInfos"`
}

// InnerGetNodeStatusResp is the response struct of node status
type InnerGetNodeStatusResp struct {
	NodeStatus string `json:"nodeStatus"`
}

type SoftwareInfo struct {
	Name            string
	Version         string
	InactiveVersion string
}

// EdgeReportSoftwareInfoReq [struct] to report edge software info
type EdgeReportSoftwareInfoReq struct {
	SerialNumber string         `json:"serialNumber"`
	SoftwareInfo []SoftwareInfo `json:"SoftwareInfo"`
}

// ProgressInfo [struct] to report edge software upgrade result info
type ProgressInfo struct {
	Progress uint64 `json:"progress"`
	Res      string `json:"res"`
	Msg      string `json:"msg"`
}

// EdgeReportUpgradeResInfoReq [struct] to report edge software upgrade progress
type EdgeReportUpgradeResInfoReq struct {
	SerialNumber string       `json:"serialNumber"`
	ProgressInfo ProgressInfo `json:"upgradeResInfo"`
}

// InnerGetNodeInfoResReq the response struct of node info
type InnerGetNodeInfoResReq struct {
	ModuleName string `json:"moduleName"`
}
