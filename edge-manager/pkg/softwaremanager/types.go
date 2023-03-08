// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package softwaremanager types
package softwaremanager

// SftAuthInfo [struct] to describe software auth info
type SftAuthInfo struct {
	UserName string  `json:"username"`
	Password *[]byte `json:"password"`
}

// UrlInfo [struct] to describe software url info
type UrlInfo struct {
	Type      string `json:"type"`
	Url       string `json:"url"`
	CreatedAt string `json:"createdAt"`
	Version   string `json:"version"`
}

// UrlUpdateInfo [struct] to describe software url update info
type UrlUpdateInfo struct {
	Option   string    `json:"option"`
	UrlInfos []UrlInfo `json:"urlInfos"`
}
