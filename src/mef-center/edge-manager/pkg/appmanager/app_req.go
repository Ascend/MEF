// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
	AppID        uint64   `json:"appID"`
	NodeGroupIds []uint64 `json:"nodeGroupIds"`
}

// UndeployAppReq Undeploy application
type UndeployAppReq struct {
	AppID        uint64   `json:"appID"`
	NodeGroupIds []uint64 `json:"nodeGroupIds"`
}

// GetAppByAppIdReq get app by application id
type GetAppByAppIdReq struct {
	AppID uint64 `json:"appID"`
}

// AppInstanceResp encapsulate app instance information for return
type AppInstanceResp struct {
	AppID         uint64              `json:"appID"`
	AppName       string              `json:"appName"`
	AppStatus     string              `json:"appStatus"`
	NodeGroupInfo types.NodeGroupInfo `json:"nodeGroupInfo"`
	NodeID        uint64              `json:"nodeID"`
	NodeName      string              `json:"nodeName"`
	NodeStatus    string              `json:"nodeStatus"`
	CreatedAt     string              `json:"createdAt"`
	ContainerInfo []ContainerInfo     `json:"containerInfo"`
}

// ContainerInfo encapsulate container details of an app instance
type ContainerInfo struct {
	Name         string `json:"name"`
	Image        string `json:"image"`
	Status       string `json:"status"`
	RestartCount int32  `json:"restartCount"`
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

// ListAppInstancesResp encapsulate app instances list for return
type ListAppInstancesResp struct {
	AppInstances []AppInstanceResp `json:"appInstances"`
	Total        int64             `json:"total"`
}
