// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager database table
package appmanager

// Container encapsulate container request
type Container struct {
	Name             string            `json:"name"`
	Image            string            `json:"image"`
	ImageVersion     string            `json:"imageVersion"`
	CpuRequest       float64           `json:"cpuRequest"`
	CpuLimit         *float64          `json:"cpuLimit,omitempty"`
	MemRequest       int64             `json:"memRequest"`
	MemLimit         *int64            `json:"memLimit,omitempty"`
	Npu              *int64            `json:"npu,omitempty"`
	Command          []string          `json:"command"`
	Args             []string          `json:"args"`
	Env              []EnvVar          `json:"env"`
	Ports            []ContainerPort   `json:"containerPort"`
	UserID           *int64            `json:"userID"`
	GroupID          *int64            `json:"groupID"`
	HostPathVolumes  []HostPathVolume  `json:"hostPathVolumes"`
	ConfigmapVolumes []ConfigmapVolume `json:"configmapVolumes"`
}

// HostPathVolume [struct] for host path
type HostPathVolume struct {
	Name      string `json:"name"`
	HostPath  string `json:"hostPath"`
	MountPath string `json:"mountPath"`
}

// ConfigmapVolume [struct] for configmap volume
type ConfigmapVolume struct {
	Name          string `json:"name"`
	ConfigmapName string `json:"configmapName"`
	MountPath     string `json:"mountPath"`
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
	HostIP        string `json:"hostIP"`
	HostPort      int32  `json:"hostPort"`
}
