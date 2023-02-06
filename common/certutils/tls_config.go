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
	rootCaPemBytes, _ := utils.LoadFile(tlsCertInfo.RootCaPath)
	if rootCaPemBytes == nil {
		return nil, fmt.Errorf("get tls failed, load root ca path [%s] file failed", tlsCertInfo.RootCaPath)
	}
	if tlsCertInfo.SvrFlag {
		pemPair, err := GetCertPairForPem(tlsCertInfo.CertPath, tlsCertInfo.KeyPath, tlsCertInfo.KmcCfg)
		if err != nil {
			return nil, err
		}
		return getTlsConfig(rootCaPemBytes, pemPair.CertPem, pemPair.KeyPem, tlsCertInfo.SvrFlag, tlsCertInfo.IgnoreCltCert)
	}
	// client may not send certs
	if tlsCertInfo.CertPath != "" && tlsCertInfo.KeyPath != "" {
		pemPair, err := GetCertPairForPem(tlsCertInfo.CertPath, tlsCertInfo.KeyPath, tlsCertInfo.KmcCfg)
		if err != nil {
			return nil, err
		}
		return getTlsConfig(rootCaPemBytes, pemPair.CertPem, pemPair.KeyPem, tlsCertInfo.SvrFlag, tlsCertInfo.IgnoreCltCert)
	}
	return getTlsConfig(rootCaPemBytes, nil, nil, tlsCertInfo.SvrFlag, tlsCertInfo.IgnoreCltCert)
}

func getTlsConfig(rootPem, certPem, keyPem []byte, svrFlag bool, ignoreCltCert bool) (*tls.Config, error) {
	rootCaPool := x509.NewCertPool()
	if ok := rootCaPool.AppendCertsFromPEM(rootPem); !ok {
		return nil, errors.New("append root ca to cert pool failed")
	}
	var pair tls.Certificate
	var err error
	if len(certPem) > 0 && len(keyPem) > 0 {
		pair, err = tls.X509KeyPair(certPem, keyPem)
	}
	if err != nil {
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
		MinVersion: tls.VersionTLS13,
	}
	if svrFlag {
		tlsCfg.ClientCAs = rootCaPool
		tlsCfg.Certificates = []tls.Certificate{pair}
		if ignoreCltCert {
			tlsCfg.ClientAuth = tls.VerifyClientCertIfGiven
		} else {
			tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
		}
	} else {
		tlsCfg.RootCAs = rootCaPool
		if len(certPem) > 0 && len(keyPem) > 0 {
			tlsCfg.Certificates = []tls.Certificate{pair}
		}
	}
	return tlsCfg, nil
}
