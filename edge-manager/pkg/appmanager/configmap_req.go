// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager request about configmap
package appmanager

// ConfigmapReq create or update configmap request
type ConfigmapReq struct {
	ConfigmapName    string             `json:"configmapName"`
	Description      string             `json:"description,omitempty"`
	ConfigmapContent []ConfigmapContent `json:"configmapContent,omitempty"`
}

// ConfigmapContent struct configmap content
type ConfigmapContent struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}

// QueryConfigmapReturnInfo query configmap return info
type QueryConfigmapReturnInfo struct {
	ConfigmapID      int64              `json:"configmapID"`
	ConfigmapName    string             `json:"configmapName"`
	Description      string             `json:"description,omitempty"`
	ConfigmapContent []ConfigmapContent `json:"configmapContent,omitempty"`
	CreatedAt        string             `json:"createdAt"`
	UpdatedAt        string             `json:"updatedAt"`
}

// DeleteConfigmapReq delete configmap request
type DeleteConfigmapReq struct {
	ConfigmapIDs []int64 `json:"configmapIDs"`
}

// ConfigmapInstanceResp encapsulate configmap instance information
type ConfigmapInstanceResp struct {
	ConfigmapID      int64              `json:"configmapID"`
	ConfigmapName    string             `json:"configmapName"`
	Description      string             `json:"description,omitempty"`
	ConfigmapContent []ConfigmapContent `json:"configmapContent,omitempty"`
	CreatedAt        string             `json:"createdAt"`
	UpdatedAt        string             `json:"updatedAt"`
}

// ListConfigmapReturnInfo encapsulate configmap list
type ListConfigmapReturnInfo struct {
	// ConfigmapInstanceResp is configmap information
	ConfigmapInstance []ConfigmapInstanceResp `json:"configmapInstance"`
	// Total is num of configmapInfos counted by search
	Total int64 `json:"total"`
}
