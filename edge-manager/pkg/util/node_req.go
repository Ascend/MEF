// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package util to init util service
package util

import (
	"encoding/json"
	"errors"
	"strconv"
)

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

// GetNodeDetailReq request object
type GetNodeDetailReq struct {
	Id int64 `json:"id"`
}

// UnmarshalJSON custom JSON unmarshal
func (req *GetNodeDetailReq) UnmarshalJSON(input []byte) error {
	objMap := make(map[string][]string)
	if err := json.Unmarshal(input, &objMap); err != nil {
		return err
	}
	idStr, ok := objMap["id"]
	if !ok || len(idStr) != 1 {
		return errors.New("bad data format")
	}
	id, err := strconv.Atoi(idStr[0])
	if err != nil {
		return err
	}
	req.Id = int64(id)
	return nil
}

// Check request validator
func (req GetNodeDetailReq) Check() error {
	return nil
}

// GetNodeDetailResp nodeDetail
type GetNodeDetailResp struct {
	Id          int64  `json:"ID"`
	NodeName    string `json:"nodeName"`
	UniqueName  string `json:"uniqueName"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	Cpu         int64  `json:"cpu"`
	Memory      int64  `json:"memory"`
	Npu         string `json:"npu"`
	NodeType    string `json:"nodeType"`
	NodeGroup   string `json:"nodeGroup"`
}

// ModifyNodeGroupReq request object
type ModifyNodeGroupReq struct {
	NodeId      int64  `json:"nodeID"`
	NodeName    string `json:"nodeName"`
	Description string `json:"description"`
}

// Check request validator
func (req ModifyNodeGroupReq) Check() error {
	return nil
}

// GetNodeStatisticsResp node statistics data
type GetNodeStatisticsResp = map[string]int64
