// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package certutils for cert utils
package certutils

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"

	"huawei.com/mindx/common/rand"
	"huawei.com/mindx/common/utils"
)

// GetTlsCfgWithPath [method] for get tls config
func GetTlsCfgWithPath(tlsCertPath TlsCertInfo) (*tls.Config, error) {
	pemPair, err := GetCertPairForPem(tlsCertPath.CertPath, tlsCertPath.KeyPath, tlsCertPath.KmcCfg)
	if err != nil {
		return nil, err
	}
	rootCaPemBytes, err := utils.LoadFile(tlsCertPath.RootCaPath)
	if rootCaPemBytes == nil {
		return nil, fmt.Errorf("get tls failed, load root ca path [%s] file failed", tlsCertPath.RootCaPath)
	}
	return getTlsConfig(rootCaPemBytes, pemPair.CertPem, pemPair.KeyPem, tlsCertPath.SvrFlag)
}

func getTlsConfig(rootPem, certPem, keyPem []byte, svrFlag bool) (*tls.Config, error) {
	rootCaPool := x509.NewCertPool()
	ok := rootCaPool.AppendCertsFromPEM(rootPem)
	if !ok {
		return nil, errors.New("append root ca to cert pool failed")
	}

	pair, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		return nil, err
	}
	tlsCfg := &tls.Config{
		Rand:               rand.Reader,
		Certificates:       []tls.Certificate{pair},
		InsecureSkipVerify: false,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		},
		MinVersion: tls.VersionTLS13,
	}
	if svrFlag {
		tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
		tlsCfg.ClientCAs = rootCaPool
	} else {
		tlsCfg.RootCAs = rootCaPool
	}
	return tlsCfg, nil
}
