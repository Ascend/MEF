// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package certutils

import (
	"crypto/rsa"
	"crypto/x509"
	"net"

	"huawei.com/mindxedge/base/common"
)

// CaPairInfo define cert and key pair info struct
type CaPairInfo struct {
	Cert   *x509.Certificate
	PriKey *rsa.PrivateKey
}

// CaPairInfoWithPem [struct] for ca pair info with pem encoded
type CaPairInfoWithPem struct {
	CertPem []byte
	KeyPem  []byte
}

// TlsCertInfo [struct] for get tls config parameters
type TlsCertInfo struct {
	KmcCfg        *common.KmcCfg
	KeyPath       string
	CertPath      string
	RootCaPath    string
	KeyContent    []byte // plain key bytes, make sure clear it after use
	CertContent   []byte
	RootCaContent []byte
	SvrFlag       bool
	RootCaOnly    bool // use root ca cert only, skip svc cert
}

// CertSan [struct] for server cert san fields
type CertSan struct {
	DnsName []string
	IpAddr  []net.IP
}

// ClientCertResp cert resp for client
type ClientCertResp struct {
	CertName     string `json:"certName"`
	CertContent  string `json:"certContent"`
	CertOpt      string `json:"certOpt" default:"update"`
	ImageAddress string `json:"imageAddress"`
}

// UpdateClientCert update cert struct
type UpdateClientCert struct {
	CertName    string `json:"certName"`
	CertContent []byte `json:"certContent"`
	CertOpt     string `json:"certOpt" default:"update"`
}
