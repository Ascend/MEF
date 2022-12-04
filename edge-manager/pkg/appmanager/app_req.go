// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init util service
package appmanager

// CreateAppReq Create application
type CreateAppReq struct {
	AppId       uint64      `json:"appId"`
	AppName     string      `json:"appName"`
	Version     string      `json:"version"`
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
	AppId uint64 `json:"appId"`
}

// DeployAppReq Deploy application
type DeployAppReq struct {
	AppId         uint64 `json:"appId"`
	NodeGroupName string `json:"nodeGroupName"`
}

// UndeployAppReq Undeploy application
type UndeployAppReq struct {
	AppId         uint64 `json:"appId"`
	NodeGroupName string `json:"nodeGroupName"`
}

// GetAppByAppIdReq get app by application id
type GetAppByAppIdReq struct {
	AppId uint64 `json:"appId"`
}

// AppInstanceResp encapsulate app instance information for return
type AppInstanceResp struct {
	AppName       string `json:"appName"`
	NodeGroupName string `json:"nodeGroupName"`
	NodeName      string `json:"nodeName"`
	NodeStatus    string `json:"nodeStatus"`
	AppStatus     string `json:"appStatus"`
}

// ListReturnInfo encapsulate app list
type ListReturnInfo struct {
	// AppInfo is app information
	AppInfo []AppReturnInfo
	// Total is num of appInfos
	Total int
}

// AppReturnInfo encapsulate app information for return
type AppReturnInfo struct {
	AppId         uint64      `json:"appId"`
	AppName       string      `json:"appName"`
	Version       string      `json:"version"`
	Description   string      `json:"description"`
	CreatedAt     string      `json:"createdAt"`
	ModifiedAt    string      `json:"modifiedAt"`
	NodeGroupName string      `json:"nodeGroupName"`
	Containers    []Container `json:"containers"`
}
