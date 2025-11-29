// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package util this file for common methods when upgrade
package util

// SoftwareDownloadInfo content for download software
type SoftwareDownloadInfo struct {
	SoftwareName string       `json:"softwareName"`
	DownloadInfo DownloadInfo `json:"downloadInfo"`
}

// DownloadInfo [struct] to software download info
type DownloadInfo struct {
	Package  string `json:"package"`
	SignFile string `json:"signFile,omitempty"`
	CrlFile  string `json:"crlFile,omitempty"`
	UserName string `json:"username"`
	Password []byte `json:"password"`
}

// SoftwareUpdateInfo content for download software
type SoftwareUpdateInfo struct {
	SoftwareName string `json:"softwareName"`
}

// ClientCertResp query cert content resp
type ClientCertResp struct {
	CertName     string `json:"certName"`
	CertContent  string `json:"certContent"`
	CrlContent   string `json:"crlContent"`
	CertOpt      string `json:"certOpt" default:"update"`
	ImageAddress string `json:"imageAddress"`
}
