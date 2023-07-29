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

// DeleteCmReq delete configmap request
type DeleteCmReq struct {
	ConfigmapIDs []uint64 `json:"configmapIDs"`
}

// ConfigmapInstance encapsulate configmap instance information
type ConfigmapInstance struct {
	ConfigmapID       uint64             `json:"configmapID"`
	ConfigmapName     string             `json:"configmapName"`
	Description       string             `json:"description,omitempty"`
	ConfigmapContent  []ConfigmapContent `json:"configmapContent,omitempty"`
	AssociatedAppNum  uint64             `json:"associatedAppNum"`
	AssociatedAppList []uint64           `json:"associatedAppList"`
	CreatedAt         string             `json:"createdAt"`
	UpdatedAt         string             `json:"updatedAt"`
}

// ListConfigmapResp encapsulate configmap list
type ListConfigmapResp struct {
	// Configmaps are configmap information
	Configmaps []ConfigmapInstance `json:"configmaps"`
	// Total is num of configmapInfos counted by search
	Total int64 `json:"total"`
}
