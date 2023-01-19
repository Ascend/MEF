// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init util service
package appmanager

import "edge-manager/pkg/types"

// CreateAppReq Create application
type CreateAppReq struct {
	AppName     string      `json:"appName"`
	Description string      `json:"description"`
	Containers  []Container `json:"containers"`
}

// UpdateAppReq update application
type UpdateAppReq struct {
	AppID uint64 `json:"appID"`
	CreateAppReq
}

// DeleteAppReq Delete application
type DeleteAppReq struct {
	AppIDs []uint64 `json:"appIDs"`
}

// DeployAppReq Deploy application
type DeployAppReq struct {
	AppID        uint64  `json:"appID"`
	NodeGroupIds []int64 `json:"nodeGroupIds"`
}

// UndeployAppReq Undeploy application
type UndeployAppReq struct {
	AppID        uint64  `json:"appID"`
	NodeGroupIds []int64 `json:"nodeGroupIds"`
}

// GetAppByAppIdReq get app by application id
type GetAppByAppIdReq struct {
	AppID uint64 `json:"appID"`
}

// AppInstanceResp encapsulate app instance information for return
type AppInstanceResp struct {
	AppName       string              `json:"appName"`
	AppStatus     string              `json:"appStatus"`
	NodeGroupInfo types.NodeGroupInfo `json:"nodeGroupInfo"`
	NodeID        int64               `json:"nodeID"`
	NodeName      string              `json:"nodeName"`
	NodeStatus    string              `json:"nodeStatus"`
	CreatedAt     string              `json:"createdAt"`
	ContainerInfo []ContainerInfo     `json:"containerInfo"`
}

// ContainerInfo encapsulate container details of an app instance
type ContainerInfo struct {
	Name   string `json:"name"`
	Image  string `json:"image"`
	Status string `json:"status"`
}

// CreateReturnInfo for create app
type CreateReturnInfo struct {
	AppID uint64 `json:"appID"`
}

// ListReturnInfo encapsulate app list
type ListReturnInfo struct {
	// AppInfo is app information
	AppInfo []AppReturnInfo `json:"appInfo"`
	// Total is num of appInfos counted by search
	Total int64 `json:"total"`
	// Deployed is num of deployed apps of all
	Deployed int64 `json:"deployed"`
	// UnDeployed is num of deployed apps of all
	UnDeployed int64 `json:"unDeployed"`
}

// AppReturnInfo encapsulate app information for return
type AppReturnInfo struct {
	AppID          uint64                `json:"appID"`
	AppName        string                `json:"appName"`
	Description    string                `json:"description"`
	CreatedAt      string                `json:"createdAt"`
	ModifiedAt     string                `json:"modifiedAt"`
	NodeGroupInfos []types.NodeGroupInfo `json:"nodeGroupInfos"`
	Containers     []Container           `json:"containers"`
}

// AppInstanceOfNodeResp encapsulate app instance information of a certain node
type AppInstanceOfNodeResp struct {
	AppName       string              `json:"appName"`
	AppStatus     string              `json:"appStatus"`
	Description   string              `json:"description"`
	CreatedAt     string              `json:"createdAt"`
	ChangedAt     string              `json:"changedAt"`
	NodeGroupInfo types.NodeGroupInfo `json:"nodeGroupInfo"`
}

// AppTemplate app template detail
type AppTemplate struct {
	Id          uint64      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	CreatedAt   string      `json:"createdAt"`
	ModifiedAt  string      `json:"modifiedAt"`
	Containers  []Container `json:"containers"`
}

// CreateTemplateReq create app template
type CreateTemplateReq struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Containers  []Container `json:"containers"`
}

// UpdateTemplateReq update app template
type UpdateTemplateReq struct {
	Id uint64 `json:"id"`
	CreateTemplateReq
}

// ListTemplatesResp encapsulate app list
type ListTemplatesResp struct {
	// AppTemplates app template info
	AppTemplates []AppTemplate `json:"appTemplates"`
	// Total is num of appInfos
	Total int64 `json:"total"`
}

// DeleteTemplateReq request body to delete app template
type DeleteTemplateReq struct {
	Ids []uint64 `json:"ids"`
}
