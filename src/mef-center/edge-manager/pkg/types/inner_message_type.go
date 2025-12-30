// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package types defines structs which could be used in different package
package types

import "k8s.io/api/core/v1"

// InnerGetNodeInfoByNameReq is the request struct for internal module to get node info by node name
type InnerGetNodeInfoByNameReq struct {
	UniqueName string `json:"uniqueName"`
}

// InnerGetSfwInfoBySNReq is the request struct for internal module to get node info by SerialNumber
type InnerGetSfwInfoBySNReq struct {
	SerialNumber string `json:"serialNumber"`
}

// InnerGetNodeGroupInfosReq is the request struct for internal module to get node group infos by group ids
type InnerGetNodeGroupInfosReq struct {
	NodeGroupIds []uint64 `json:"nodeGroupIDs"`
}

// InnerGetNodeInfosReq is the request struct for internal module to get node infos by node ids
type InnerGetNodeInfosReq struct {
	NodeIds []uint64 `json:"nodeIDs"`
}

// InnerGetNodeStatusReq is request struct for internal module to get node status
type InnerGetNodeStatusReq struct {
	UniqueName string `json:"uniqueName"`
}

// InnerGetNodesReq [struct] for getting nodes id request
type InnerGetNodesReq struct {
	NodeGroupID uint64 `json:"nodeGroupID"`
}

// InnerUpdateNodeResReq [struct] for updating node resource request
type InnerUpdateNodeResReq struct {
	NodeGroupID  uint64 `json:"nodeGroupID"`
	ResourceReqs v1.ResourceList
	IsUndeploy   bool
}

// InnerCheckNodeResReq [struct] for checking node resource request
type InnerCheckNodeResReq struct {
	NodeGroupID  uint64 `json:"nodeGroupID"`
	ResourceReqs v1.ResourceList
}

// InnerSoftwareInfoResp is the response struct of node info
type InnerSoftwareInfoResp struct {
	SoftwareInfo []SoftwareInfo `json:"softwareInfo"`
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

// InnerGetNodeInfosResp is the response struct of node group infos by group ids
type InnerGetNodeInfosResp struct {
	NodeInfos []NodeInfo `json:"nodeInfos"`
}

// InnerGetNodeStatusResp is the response struct of node status
type InnerGetNodeStatusResp struct {
	NodeStatus string `json:"nodeStatus"`
}

// SoftwareInfo [struct] to record software info
type SoftwareInfo struct {
	Name            string
	Version         string
	InactiveVersion string
}

// EdgeReportSoftwareInfoReq [struct] to report edge software info
type EdgeReportSoftwareInfoReq struct {
	SerialNumber string         `json:"serialNumber"`
	SoftwareInfo []SoftwareInfo `json:"softwareInfo"`
}

// ProgressInfo [struct] to report edge software download result info
type ProgressInfo struct {
	Progress uint64 `json:"progress"`
	Res      string `json:"res"`
	Msg      string `json:"msg"`
}

// EdgeDownloadResInfo [struct] to report edge software download result info
type EdgeDownloadResInfo struct {
	SerialNumber string       `json:"serialNumber"`
	ProgressInfo ProgressInfo `json:"upgradeResInfo"`
}

// InnerGetNodesResp [struct] for getting nodes id response
type InnerGetNodesResp struct {
	NodeIDs []uint64 `json:"nodeIDs"`
}

// InnerGetNodeInfoResReq the response struct of node info
type InnerGetNodeInfoResReq struct {
	ModuleName string `json:"moduleName"`
}
