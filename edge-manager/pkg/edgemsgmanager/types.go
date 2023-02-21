// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager to deal node msg
package edgemsgmanager

// SoftwareDownloadInfo content for download software
type SoftwareDownloadInfo struct {
	NodeIDs         []uint64     `json:"nodeIDs"`
	SerialNumbers   []string     `json:"serialNumbers"`
	SoftwareName    string       `json:"softwareName"`
	SoftwareVersion string       `json:"softwareVersion,omitempty"`
	DownloadInfo    DownloadInfo `json:"downloadInfo"`
}

// DownloadInfo [struct] to software download info
type DownloadInfo struct {
	Package  string  `json:"package"`
	SignFile *string `json:"signFile,omitempty"`
	CrlFile  *string `json:"crlFile,omitempty"`
	UserName string  `json:"username"`
	Password []byte  `json:"password"`
}

// upgradeInfoReq effect software
type upgradeInfoReq struct {
	NodeIDs       []uint64 `json:"nodeIDs"`
	SerialNumbers []string `json:"serialNumbers"`
	SoftwareName  string   `json:"softwareName"`
}
