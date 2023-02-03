// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager to deal node msg
package edgemsgmanager

// EdgeUpgradeInfoReq software upgrade req
type EdgeUpgradeInfoReq struct {
	NodeIDs         []uint64     `json:"nodeIDs"`
	SNs             []string     `json:"sns"`
	SoftWareName    string       `json:"softWareName"`
	SoftWareVersion string       `json:"softWarVersion"`
	DownloadInfo    DownloadInfo `json:"downloadInfo"`
}

// DownloadInfo [struct] for package download info
type DownloadInfo struct {
	Url      string `json:"url"`
	UserName string `json:"userName"`
	Password string `json:"password"`
}

// EffectInfoReq effect software
type EffectInfoReq struct {
	NodeIDs []uint64 `json:"nodeIDs"`
	SNs     []string `json:"sns"`
}
