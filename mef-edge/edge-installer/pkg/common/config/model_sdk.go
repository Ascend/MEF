// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package config for
package config

// DomainConfig signature config struct
type DomainConfig struct {
	Domain string
	IP     string
}

// DomainConfigs slice of signature config struct
type DomainConfigs struct {
	Configs []DomainConfig
}

// NetManager net manager struct
type NetManager struct {
	NetType  string
	IP       string
	Port     int
	AuthPort int
	WithOm   bool
	Token    []byte
}

// ImageConfig image config struct
type ImageConfig struct {
	ImageAddress string
}

// ProgressInfo [struct] to report edge software upgrade result info
type ProgressInfo struct {
	Progress uint64 `json:"progress"`
	Res      string `json:"res"`
	Msg      string `json:"msg"`
}

// CertReq [struct] request cert from edge-om
type CertReq struct {
	CertName string
}

// CertResp [struct] is the response of cert request
type CertResp struct {
	CertReq
	CertContent []byte
	CrlContent  []byte
	ErrorMsg    string
}

// DirReq [struct] request for prepare directory
type DirReq struct {
	Path     string
	ToDelete bool
}
