// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package util common struct in edge-manager
package util

// DealSfwContent deal software content
type DealSfwContent struct {
	NodeId          string `json:"nodeId"`
	Url             string `json:"url"`
	SoftwareName    string `json:"softwareName,omitempty"`
	SoftwareVersion string `json:"softwareVersion,omitempty"`
	Username        string `json:"username"`
	Password        string `json:"password"`
}
