// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package util to init util service
package util

// UpgradeSfwReq request of upgrading software
type UpgradeSfwReq struct {
	NodeIDs         []int64 `json:"nodeIDs"`
	SoftwareName    string  `json:"softwareName"`
	SoftwareVersion string  `json:"softwareVersion"`
}

// DownloadSfwReq request of downloading software
type DownloadSfwReq struct {
	NodeID          string `json:"nodeID"`
	SoftwareName    string `json:"softwareName"`
	SoftwareVersion string `json:"softwareVersion,omitempty"`
}
