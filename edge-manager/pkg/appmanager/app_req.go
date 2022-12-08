// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init util service
package appmanager

import "edge-manager/pkg/nodemanager"

// CreateAppReq Create application
type CreateAppReq struct {
	AppId       uint64      `json:"appId"`
	AppName     string      `json:"appName"`
	Description string      `json:"description"`
	Containers  []Container `json:"containers"`
}

// Container encapsulate container request
type Container struct {
	Name         string          `json:"name"`
	Image        string          `json:"image"`
	ImageVersion string          `json:"imageVersion"`
	CpuRequest   string          `json:"cpuRequest"`
	CpuLimit     string          `json:"cpuLimit"`
	MemRequest   string          `json:"memRequest"`
	MemLimit     string          `json:"memLimit"`
	Npu          string          `json:"npu"`
	Command      []string        `json:"command"`
	Args         []string        `json:"args"`
	Env          []EnvVar        `json:"env"`
	Ports        []ContainerPort `json:"containerPort"`
	UserId       int             `json:"userId"`
	GroupId      int             `json:"groupId"`
}

// EnvVar encapsulate env request
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ContainerPort provide ports mapping
type ContainerPort struct {
	Name          string `json:"name"`
	Proto         string `json:"proto"`
	ContainerPort int32  `json:"containerPort"`
	HostIp        string `json:"hostIp"`
	HostPort      int32  `json:"hostPort"`
}

// DeleteAppReq Delete application
type DeleteAppReq struct {
	AppIdList []uint64 `json:"appId"`
}

// DeployAppReq Deploy application
type DeployAppReq struct {
	AppId         uint64          `json:"appId"`
	NodeGroupInfo []NodeGroupInfo `json:"nodeGroupInfo"`
}

// UndeployAppReq Undeploy application
type UndeployAppReq struct {
	AppId    uint64     `json:"appId"`
	NodeInfo []NodeInfo `json:"nodeInfo"`
}

// NodeInfo get node info
type NodeInfo struct {
	NodeID        int64  `json:"nodeID"`
	NodeGroupName string `json:"nodeGroupName"`
}

// NodeGroupInfo get group info
type NodeGroupInfo struct {
	NodeGroupID   int64  `json:"nodeGroupID"`
	NodeGroupName string `json:"nodeGroupName"`
}

// GetAppByAppIdReq get app by application id
type GetAppByAppIdReq struct {
	AppId uint64 `json:"appId"`
}

// AppInstanceResp encapsulate app instance information for return
type AppInstanceResp struct {
	AppName       string          `json:"appName"`
	NodeGroupId   int64           `json:"nodeGroupId"`
	NodeGroupName string          `json:"nodeGroupName"`
	NodeId        int64           `json:"nodeId"`
	NodeName      string          `json:"nodeName"`
	NodeStatus    string          `json:"nodeStatus"`
	AppStatus     string          `json:"appStatus"`
	CreatedAt     string          `json:"createdAt"`
	ContainerInfo []ContainerInfo `json:"containerInfo"`
}

// ContainerInfo encapsulate container details of an app instance
type ContainerInfo struct {
	Name   string `json:"name"`
	Image  string `json:"image"`
	Status string `json:"status"`
}

// AppInstanceInfo encapsulate app instance information
type AppInstanceInfo struct {
	// AppInfo is app information
	AppInfo AppInfo `json:"appInfo"`
	// NodeGroup is node group information of app
	NodeGroup nodemanager.NodeGroup `json:"nodeGroup"`
}

// CreateReturnInfo for create app
type CreateReturnInfo struct {
	AppId uint64 `json:"appId"`
}

// ListReturnInfo encapsulate app list
type ListReturnInfo struct {
	// AppInfo is app information
	AppInfo []AppReturnInfo `json:"appInfo"`
	// Total is num of appInfos
	Total int64 `json:"total"`
}

// AppReturnInfo encapsulate app information for return
type AppReturnInfo struct {
	AppId         uint64      `json:"appId"`
	AppName       string      `json:"appName"`
	Description   string      `json:"description"`
	CreatedAt     string      `json:"createdAt"`
	ModifiedAt    string      `json:"modifiedAt"`
	NodeGroupName string      `json:"nodeGroupName"`
	NodeGroupId   []int64     `json:"nodeGroupId"`
	Containers    []Container `json:"containers"`
}

// AppInstanceOfNodeResp encapsulate app instance information of a certain node
type AppInstanceOfNodeResp struct {
	AppName       string `json:"appName"`
	AppStatus     string `json:"appStatus"`
	Description   string `json:"description"`
	CreatedAt     string `json:"createdAt"`
	ChangedAt     string `json:"changedAt"`
	NodeGroupName string `json:"nodeGroupName"`
	NodeGroupID   int64  `json:"nodeGroupID"`
}
