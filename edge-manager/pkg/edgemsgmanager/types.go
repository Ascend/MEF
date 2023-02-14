// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager to deal node msg
package edgemsgmanager

// SoftwareDownloadInfo content for download software
type SoftwareDownloadInfo struct {
	NodeIDs         []uint64    `json:"nodeIDs"`
	SerialNumbers   []string    `json:"serialNumber"`
	SoftwareName    string      `json:"softwareName"`
	SoftwareVersion string      `json:"softwareVersion,omitempty"`
	HttpsServer     HttpsServer `json:"httpsServer"`
}

type HttpsServer struct {
	Package  string `json:"package"`
	SignFile string `json:"signFile"`
	CrlFile  string `json:"crlFile,omitempty"`
	UserName string `json:"username"`
	Password string `json:"password"`
}

// EffectInfoReq effect software
type EffectInfoReq struct {
	NodeIDs     []uint64 `json:"nodeIDs"`
	SerialNames []string `json:"SerialNames"`
}
