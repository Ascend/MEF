// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package appmanager to init app manager database table
package appmanager

// Container encapsulate container request
type Container struct {
	Name            string           `json:"name"`
	Image           string           `json:"image"`
	ImageVersion    string           `json:"imageVersion"`
	CpuRequest      float64          `json:"cpuRequest"`
	CpuLimit        *float64         `json:"cpuLimit,omitempty"`
	MemRequest      int64            `json:"memRequest"`
	MemLimit        *int64           `json:"memLimit,omitempty"`
	Npu             *int64           `json:"npu,omitempty"`
	Command         []string         `json:"command"`
	Args            []string         `json:"args"`
	Env             []EnvVar         `json:"env"`
	Ports           []ContainerPort  `json:"containerPort"`
	UserID          *int64           `json:"userID"`
	GroupID         *int64           `json:"groupID"`
	HostPathVolumes []HostPathVolume `json:"hostPathVolumes"`
}

// HostPathVolume [struct] for host path
type HostPathVolume struct {
	Name      string `json:"name"`
	HostPath  string `json:"hostPath"`
	MountPath string `json:"mountPath"`
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
