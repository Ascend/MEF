// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package util to init util service
package util

// CreateAppReq Create application
type CreateAppReq struct {
	AppName     string         `json:"appName"`
	Version     string         `json:"version"`
	Description string         `json:"description"`
	Containers  []ContainerReq `json:"containers"`
}

// ContainerReq encapsulate container request
type ContainerReq struct {
	ContainerName string         `json:"containerName"`
	CpuRequest    string         `json:"cpuRequest"`
	CpuLimit      string         `json:"cpuLimit"`
	MemRequest    string         `json:"memRequest"`
	MemLimit      string         `json:"memLimit"`
	Npu           string         `json:"npu"`
	ImageName     string         `json:"imageName"`
	ImageVersion  string         `json:"imageVersion"`
	Command       []string       `json:"command"`
	Env           []EnvReq       `json:"env"`
	ContainerPort []PortTransfer `json:"containerPort"`
	UserId        int            `json:"userId"`
	GroupId       int            `json:"groupId"`
}

// EnvReq encapsulate env request
type EnvReq struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PortTransfer provide ports mapping
type PortTransfer struct {
	Name          string `json:"name"`
	Proto         string `json:"proto"`
	ContainerPort int32  `json:"containerPort"`
	HostIp        string `json:"hostIp"`
	HostPort      int32  `json:"hostPort"`
}

// UpdateAppReq Update application
type UpdateAppReq struct {
	AppID     uint64 `json:"appID"`
	ImageName string `json:"imageName"`
}

// DeleteAppReq Delete application
type DeleteAppReq struct {
	AppName string `json:"appName"`
}

// DeployAppReq Deploy application
type DeployAppReq struct {
	AppName       string `json:"appName"`
	NodeGroupName string `json:"nodeGroupName"`
}

// UndeployAppReq Undeploy application
type UndeployAppReq struct {
	AppName       string `json:"appName"`
	NodeGroupName string `json:"nodeGroupName"`
}

// GetAppByAppIdReq get app by application id
type GetAppByAppIdReq struct {
	AppId int `json:"appId"`
}
