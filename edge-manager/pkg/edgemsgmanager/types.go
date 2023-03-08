// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager to deal node msg
package edgemsgmanager

import "edge-manager/pkg/types"

// SoftwareDownloadInfo content for download software
type SoftwareDownloadInfo struct {
	SerialNumbers []string           `json:"serialNumbers"`
	SoftwareName  string             `json:"softwareName"`
	DownloadInfo  types.DownloadInfo `json:"downloadInfo"`
}

// UpdateInfoReq update software
type UpdateInfoReq struct {
	SerialNumbers []string `json:"serialNumbers"`
	SoftwareName  string   `json:"softwareName"`
}
