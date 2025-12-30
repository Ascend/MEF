// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
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
