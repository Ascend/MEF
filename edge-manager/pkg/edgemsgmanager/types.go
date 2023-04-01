// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager to deal node msg
package edgemsgmanager

// SoftwareDownloadInfo content for download software
type SoftwareDownloadInfo struct {
	SerialNumbers []string     `json:"serialNumbers"`
	SoftwareName  string       `json:"softwareName"`
	DownloadInfo  DownloadInfo `json:"downloadInfo"`
}

// DownloadInfo [struct] to software download info
type DownloadInfo struct {
	Package  string  `json:"package"`
	SignFile string  `json:"signFile,omitempty"`
	CrlFile  string  `json:"crlFile,omitempty"`
	UserName string  `json:"username"`
	Password *[]byte `json:"password"`
}

// UpdateInfoReq update software
type UpdateInfoReq struct {
	SerialNumbers []string `json:"serialNumbers"`
	SoftwareName  string   `json:"softwareName"`
}
