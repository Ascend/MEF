// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
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
