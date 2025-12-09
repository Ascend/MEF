// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package config this file for config model define
package config

// InstallerConfig installer config struct
type InstallerConfig struct {
	InstallDir    string
	LogPath       string
	LogBackupPath string
	SerialNumber  string
}

// PodConfig pod config struct for check pod config
type PodConfig struct {
	PodSecurityConfig
	ContainerConfig
}

// PodSecurityConfig security config struct for high risk switch check of pod spec
type PodSecurityConfig struct {
	HostPid                  bool `json:"hostPid"`
	Capability               bool `json:"capability"`
	Privileged               bool `json:"privileged"`
	AllowPrivilegeEscalation bool `json:"allowPrivilegeEscalation"`
	RunAsRoot                bool `json:"runAsRoot"`
	EmptyDirVolume           bool `json:"emptyDirVolume"`
	UseHostNetwork           bool `json:"useHostNetwork"`
	UseSeccomp               bool `json:"useSeccomp"`
	UseDefaultContainerCap   bool `json:"useDefaultContainerCap"`
	AllowReadWriteRootFs     bool `json:"allowReadWriteRootFs"`
}

// ContainerConfig container config struct for spec check of pod container
type ContainerConfig struct {
	HostPath                  []string `json:"hostPath"`
	MaxContainerNumber        int      `json:"maxContainerNumber"`
	ContainerModelFileNumber  int      `json:"containerModelFileNumber"`
	TotalModelFileNumber      int      `json:"totalModelFileNumber"`
	SystemReservedCPUQuota    float64  `json:"systemReservedCPUQuota"`
	SystemReservedMemoryQuota int64    `json:"systemReservedMemoryQuota"` // unit is MB
}

// StaticInfo static info
type StaticInfo struct {
	ProductCapabilityEdge []string `json:"product_capability_edge"`
}

// Capability capability
type Capability struct {
	ProductCapability []string `json:"product_capability"`
}

// ProgressTip progress
type ProgressTip struct {
	Topic      string `json:"topic"`
	Percentage string `json:"percentage"`
	Result     string `json:"result"`
	Reason     string `json:"reason"`
}

// AlarmCertCfg alarm cert config
type AlarmCertCfg struct {
	CheckPeriod      int `json:"checkPeriod"`
	OverdueThreshold int `json:"overdueThreshold"`
}
