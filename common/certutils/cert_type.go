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
	SvrFlag       bool
	IgnoreCltCert bool
}

// CertSan [struct] for server cert san fields
type CertSan struct {
	DnsName []string
	IpAddr  []net.IP
}

// QueryCertRes query cert content res
type QueryCertRes struct {
	CertName string `json:"certName"`
	Cert     string `json:"cert"`
	Address  string `json:"address"`
}
