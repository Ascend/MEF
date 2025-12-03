// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package types

// NodeGroupInfo define node group info
type NodeGroupInfo struct {
	NodeGroupID   uint64 `json:"nodeGroupID"`
	NodeGroupName string `json:"nodeGroupName"`
}

// NodeInfo define node info
type NodeInfo struct {
	NodeID       uint64 `json:"nodeID"`
	UniqueName   string `json:"uniqueName"`
	SerialNumber string `json:"serialNumber"`
	Ip           string `json:"ip"`
}

// ListReq for common list request, PageNum and PageSize for slice page, Name for fuzzy query
type ListReq struct {
	PageNum  uint64
	PageSize uint64
	Name     string
}

// BatchResp batch request deal result
type BatchResp struct {
	SuccessIDs  []interface{}     `json:"successIDs"`
	FailedInfos map[string]string `json:"failedInfos"`
}
