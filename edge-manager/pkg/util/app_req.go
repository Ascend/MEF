// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package util to init util service
package util

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
