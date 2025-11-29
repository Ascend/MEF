// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package certutils

import (
	"crypto/rsa"
	"crypto/x509"
	"net"

	"huawei.com/mindx/common/kmc"
	hwx509 "huawei.com/mindx/common/x509"
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
	KmcCfg        *kmc.SubConfig
	KeyContent    []byte // plain key bytes, make sure clear it after use
	CertContent   []byte
	RootCaContent []byte
	CrlContent    []byte
	CrlPath       string
	KeyPath       string
	CertPath      string
	RootCaPath    string
	SvrFlag       bool
	IgnoreCltCert bool
	RootCaOnly    bool // use root ca cert only, skip svc cert
	WithBackup    bool // if true, all files from path will enable back up and restore
}

// CertSan [struct] for server cert san fields
type CertSan struct {
	DnsName []string
	IpAddr  []net.IP
}

// ClientCertResp cert resp for client
type ClientCertResp struct {
	CertName     string `json:"certName"`
	CrlContent   string `json:"crlContent"`
	CertContent  string `json:"certContent"`
	CertOpt      string `json:"certOpt"`
	ImageAddress string `json:"imageAddress"`
}

// UpdateClientCert update cert struct
type UpdateClientCert struct {
	CertContent []byte `json:"certContent"`
	CrlContent  []byte `json:"crlContent"`
	CertName    string `json:"certName"`
	CertOpt     string `json:"certOpt"`
}

// GetCrlData Get CRL data from TlsCertInfo
func (tls *TlsCertInfo) GetCrlData() *hwx509.CrlData {
	return &hwx509.CrlData{
		CrlContent: tls.CrlContent,
		CrlPath:    tls.CrlPath,
	}
}
