// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package util common struct in edge-manager
package util

// DealSfwContent deal software content
type DealSfwContent struct {
	NodeId   string `json:"nodeId"`
	Url      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
	Token    []byte `json:"token,omitempty"`
}
