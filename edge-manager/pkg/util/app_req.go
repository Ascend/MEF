// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package util to init util service
package util

// CreateAppReq Create application
type CreateAppReq struct {
	AppId       uint64         `json:"appId"`
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
	Args          []string       `json:"args"`
	Env           []EnvReq       `json:"env"`
	ContainerPort []PortTransfer `json:"containerPort"`
	UserId        int            `json:"userId"`
	GroupId       int            `json:"groupId"`
}

// EnvReq encapsulate env request
type EnvReq struct {
	Name  string `json:"name"`
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

// DeleteAppReq Delete application
type DeleteAppReq struct {
	AppIdList []uint64 `json:"appIdList"`
}

// DeployAppReq Deploy application
type DeployAppReq struct {
	AppId         uint64          `json:"appId"`
	NodeGroupInfo []NodeGroupInfo `json:"nodeGroupInfo"`
}

// UndeployAppReq Undeploy application
type UndeployAppReq struct {
	AppId         uint64          `json:"appId"`
	NodeGroupInfo []NodeGroupInfo `json:"nodeGroupInfo"`
}

// NodeGroupInfo get group Info
type NodeGroupInfo struct {
	NodeGroupID   int64  `json:"nodeGroupID"`
	NodeGroupName string `json:"nodeGroupName"`
}

// GetAppByAppIdReq get app by application id
type GetAppByAppIdReq struct {
	AppId uint64 `json:"appId"`
}
