// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package util to init util service
package util

// UpgradeSfwReq request of upgrading software
type UpgradeSfwReq struct {
	NodeId  []string `json:"node_id"`
	Name    string   `json:"software_name"`
	Version string   `json:"software_version"`
}
