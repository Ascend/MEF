// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package types for
package types

import (
	"fmt"
	"time"
)

// ListAlarmOrEventReq can not have both GroupId and NodeId
type ListAlarmOrEventReq struct {
	PageNum  uint64 `json:"pageNum"`
	PageSize uint64 `json:"pageSize"`
	Sn       string `json:"serialNumber,omitempty"`
	GroupId  uint64 `json:"groupId,omitempty"`
	IfCenter string `json:"ifCenter,omitempty"`
}

// NodeGroupDetailFromEdgeManager should be exactly same as NodeGroupDetail in edge-manager/node-manager for marshall
type NodeGroupDetailFromEdgeManager struct {
	Nodes []NodeInfo `json:"nodes"`
}

// NodeInfo query results from edge-manager
type NodeInfo struct {
	Sn string `json:"serialNumber"`
}

// AlarmBriefInfo the simple information for respond to User
type AlarmBriefInfo struct {
	ID        uint64    `json:"ID"`
	Sn        string    `json:"SerialNumber"`
	Severity  string    `json:"Severity"`
	Resource  string    `json:"Resource"`
	CreatedAt time.Time `json:"CreateAt"`
	AlarmType string    `json:"AlarmType"`
}

func (dig AlarmBriefInfo) String() string {
	return fmt.Sprintf("DigestInfo  ID: %d,SerialNumber:%s,Severity:%s,Resource:%s,CreateTime:%v", dig.ID,
		dig.Sn, dig.Severity, dig.Resource, dig.CreatedAt)
}
