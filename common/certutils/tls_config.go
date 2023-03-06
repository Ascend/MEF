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
func GetTlsCfgWithPath(tlsCertInfo TlsCertInfo) (*tls.Config, error) {
	var err error
	rootCaPemBytes := tlsCertInfo.RootCaContent
	if rootCaPemBytes == nil {
		rootCaPemBytes, err = utils.LoadFile(tlsCertInfo.RootCaPath)
		if err != nil || rootCaPemBytes == nil {
			return nil, fmt.Errorf("get tls failed, load root ca path [%s] file failed", tlsCertInfo.RootCaPath)
		}
	}
	if tlsCertInfo.RootCaOnly {
		return getTlsCfgRootCaOnly(rootCaPemBytes)
	}
	pemPair := &CaPairInfoWithPem{
		CertPem: tlsCertInfo.CertContent,
		KeyPem:  tlsCertInfo.KeyContent,
	}

	if pemPair.KeyPem == nil || pemPair.CertPem == nil {
		pemPair, err = GetCertPairForPem(tlsCertInfo.CertPath, tlsCertInfo.KeyPath, tlsCertInfo.KmcCfg)
		if err != nil {
			return nil, err
		}
	}
	return getTlsConfig(rootCaPemBytes, pemPair.CertPem, pemPair.KeyPem, tlsCertInfo.SvrFlag)
}

func getTlsConfig(rootPem, certPem, keyPem []byte, svrFlag bool) (*tls.Config, error) {
	rootCaPool := x509.NewCertPool()
	if ok := rootCaPool.AppendCertsFromPEM(rootPem); !ok {
		return nil, errors.New("append root ca to cert pool failed")
	}
	var pair tls.Certificate
	var err error
	if pair, err = tls.X509KeyPair(certPem, keyPem); err != nil {
		return nil, err
	}
	tlsCfg := &tls.Config{
		Rand:               rand.Reader,
		InsecureSkipVerify: false,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		},
		Certificates: []tls.Certificate{pair},
		MinVersion:   tls.VersionTLS13,
	}
	if svrFlag {
		tlsCfg.ClientCAs = rootCaPool
		tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
	} else {
		tlsCfg.RootCAs = rootCaPool
	}
	return tlsCfg, nil
}

func getTlsCfgRootCaOnly(rootPem []byte) (*tls.Config, error) {
	rootCaPool := x509.NewCertPool()
	if ok := rootCaPool.AppendCertsFromPEM(rootPem); !ok {
		return nil, errors.New("append root ca to cert pool failed")
	}
	return &tls.Config{
		Rand:               rand.Reader,
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
		RootCAs:    rootCaPool,
	}, nil
}
