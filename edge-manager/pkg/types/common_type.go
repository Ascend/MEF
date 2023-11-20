// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

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
