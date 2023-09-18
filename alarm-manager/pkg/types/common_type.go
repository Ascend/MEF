// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

package types

// NodeGroupInfo define node group info
type NodeGroupInfo struct {
	NodeGroupID   uint64 `json:"nodeGroupID"`
	NodeGroupName string `json:"nodeGroupName"`
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

// ListAlarmsResp return list of resp for list alarms
type ListAlarmsResp struct {
	// AppTemplates app template info
	Records []AlarmBriefInfo `json:"records"`
	// Total is num of appInfos
	Total int64 `json:"total"`
}
