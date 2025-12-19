// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package certutils for cert utils
package certutils

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/rand"
	"huawei.com/mindx/common/utils"
	hwx509 "huawei.com/mindx/common/x509"
)

// DefaultSafeCipherSuites default safe suites
var DefaultSafeCipherSuites = []uint16{
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
}

type tlsParam struct {
	rootPem       []byte
	certPem       []byte
	keyPem        []byte
	certPath      string
	keyPath       string
	KmcCfg        *kmc.SubConfig
	crls          []*pkix.CertificateList
	svrFlag       bool
	ignoreCltCert bool
}

type certFilesChecker func(tlsCertInfo TlsCertInfo) error

// GetTlsCfgWithPath [method] for get tls config
func GetTlsCfgWithPath(tlsCertInfo TlsCertInfo) (*tls.Config, error) {
	var (
		tlsConfig *tls.Config
		err       error
	)
	if tlsCertInfo.WithBackup {
		tlsConfig, err = getTlsCfgWithPathAndBackup(tlsCertInfo)
	} else {
		tlsConfig, err = getTlsCfgWithPathWithoutBackup(tlsCertInfo)
	}
	if err != nil {
		return nil, err
	}

	if len(tlsCertInfo.RootCaContent) > 0 {
		if err := hwx509.CheckPemCertPool(tlsCertInfo.RootCaContent); err != nil {
			return nil, fmt.Errorf("check root ca cert data failed: %v", err)
		}
	}
	if len(tlsCertInfo.RootCaPath) > 0 {
		if _, err := hwx509.CheckCertsChainReturnContent(tlsCertInfo.RootCaPath); err != nil {
			return nil, fmt.Errorf("check root ca cert data failed: %v", err)
		}
	}
	return tlsConfig, nil
}

// GetTlsCfgWithPathIgnoreLength [method] for get tls config ignore 2048 length error.
// This function can only be used when work with third party software.
func GetTlsCfgWithPathIgnoreLength(tlsCertInfo TlsCertInfo) (*tls.Config, error) {
	var (
		tlsConfig *tls.Config
		err       error
	)
	if tlsCertInfo.WithBackup {
		tlsConfig, err = getTlsCfgWithPathAndBackup(tlsCertInfo)
	} else {
		tlsConfig, err = getTlsCfgWithPathWithoutBackup(tlsCertInfo)
	}
	if err != nil {
		return nil, err
	}

	chainOpts := hwx509.CaChainCheckOptions{AllowMiddleStrengthRsaPublicKey: true}
	if len(tlsCertInfo.RootCaContent) > 0 {
		if err := hwx509.CheckPemCertPool(tlsCertInfo.RootCaContent, chainOpts); err != nil {
			return nil, fmt.Errorf("check root ca cert data failed: %v", err)
		}
	}
	if len(tlsCertInfo.RootCaPath) > 0 {
		if _, err := hwx509.CheckCertsChainReturnContent(tlsCertInfo.RootCaPath, chainOpts); err != nil {
			return nil, fmt.Errorf("root ca cert data failed: %v", err)
		}
	}
	return tlsConfig, nil
}

func getTlsCfgWithPathAndBackup(tlsCertInfo TlsCertInfo) (*tls.Config, error) {
	var config *tls.Config
	var err error
	// pathCheckers only check integrity of tls-related files.
	// You should verify these files to ensure the files fulfill the security requirements.
	pathCheckers := map[string]certFilesChecker{
		tlsCertInfo.RootCaPath: checkTlsRootCa,
		tlsCertInfo.CrlPath:    checkTlsCrl,
		tlsCertInfo.CertPath:   checkTlsCert,
		tlsCertInfo.KeyPath:    checkTlsKey,
	}

	config, err = getTlsCfgWithPathWithoutBackup(tlsCertInfo)
	if err == nil {
		for filePath := range pathCheckers {
			if filePath == "" {
				continue
			}
			if backupErr := backuputils.BackUpFiles(filePath); backupErr != nil {
				hwlog.RunLog.Warnf("backup file [%s] failed, error: %v", filePath, backupErr)
			}
		}
		return config, nil
	}

	// exclude case which getting tls without files path
	if tlsCertInfo.RootCaPath == "" && tlsCertInfo.CrlPath == "" &&
		tlsCertInfo.CertPath == "" && tlsCertInfo.KeyPath == "" {
		return nil, err
	}

	// try restore cert files from backup
	hwlog.RunLog.Warnf("get tls config with file failed, %v, try restore tls certs", err)
	for filePath, fn := range pathCheckers {
		// when filePath is empty, it means data is provided by []byte, no need to backup or restore.
		if filePath == "" {
			if err := fn(tlsCertInfo); err != nil {
				hwlog.RunLog.Errorf("check data failed: %v", err)
				return nil, err
			}
			continue
		}
		checkErr := fn(tlsCertInfo)
		if checkErr == nil {
			if backupErr := backuputils.BackUpFiles(filePath); backupErr != nil {
				hwlog.RunLog.Warnf("backup file [%s] failed, error: %v", filePath, backupErr)
			}
			continue
		}
		hwlog.RunLog.Warnf("check file [%s] failed, error: %v, will restore it from backup,", filePath, checkErr)
		if restoreErr := backuputils.RestoreFiles(filePath); restoreErr != nil {
			hwlog.RunLog.Errorf("restore file [%s] failed, error: %v", filePath, restoreErr)
			return nil, err
		}
	}
	return getTlsCfgWithPathWithoutBackup(tlsCertInfo)
}

func getTlsCfgWithPathWithoutBackup(tlsCertInfo TlsCertInfo) (*tls.Config, error) {
	var err error
	rootCaPemBytes := tlsCertInfo.RootCaContent
	if rootCaPemBytes == nil {
		rootCaPemBytes, err = fileutils.LoadFile(tlsCertInfo.RootCaPath)
		if err != nil || rootCaPemBytes == nil {
			return nil, fmt.Errorf("get tls failed, load root ca path [%s] file failed", tlsCertInfo.RootCaPath)
		}
	}
	var crls []*pkix.CertificateList

	if tlsCertInfo.GetCrlData().ContainsData() {
		crls, err = hwx509.ParseCrls(tlsCertInfo.GetCrlData())
		if err != nil {
			return nil, err
		}
	}

	if tlsCertInfo.RootCaOnly {
		return getTlsCfgRootCaOnly(rootCaPemBytes, crls)
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
	return getTlsConfig(tlsParam{
		rootPem:       rootCaPemBytes,
		certPem:       pemPair.CertPem,
		keyPem:        pemPair.KeyPem,
		certPath:      tlsCertInfo.CertPath,
		keyPath:       tlsCertInfo.KeyPath,
		KmcCfg:        tlsCertInfo.KmcCfg,
		svrFlag:       tlsCertInfo.SvrFlag,
		ignoreCltCert: tlsCertInfo.IgnoreCltCert,
		crls:          crls,
	})
}

func getTlsConfig(params tlsParam) (*tls.Config, error) {
	defer utils.ClearSliceByteMemory(params.keyPem)
	rootCaPool := x509.NewCertPool()
	if ok := rootCaPool.AppendCertsFromPEM(params.rootPem); !ok {
		return nil, errors.New("append root ca to cert pool failed")
	}

	tlsCfg := &tls.Config{
		Rand:               rand.Reader,
		InsecureSkipVerify: false,
		CipherSuites:       DefaultSafeCipherSuites,
		MinVersion:         tls.VersionTLS13,
	}
	tlsCfg.VerifyPeerCertificate = func(certificates [][]byte, verifiedChains [][]*x509.Certificate) error {
		for _, crl := range params.crls {
			if err := checkCaFileInCrl(verifiedChains, crl); err != nil {
				return err
			}
		}
		return nil
	}
	if params.svrFlag {
		tlsCfg.GetCertificate = func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			return getCertPairFromParam(params)
		}
		tlsCfg.ClientCAs = rootCaPool
		if params.ignoreCltCert {
			tlsCfg.ClientAuth = tls.VerifyClientCertIfGiven
		} else {
			tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
		}
	} else {
		tlsCfg.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			return getCertPairFromParam(params)
		}
		tlsCfg.RootCAs = rootCaPool
	}
	return tlsCfg, nil
}

func getTlsCfgRootCaOnly(rootPem []byte, crls []*pkix.CertificateList) (*tls.Config, error) {
	rootCaPool := x509.NewCertPool()
	if ok := rootCaPool.AppendCertsFromPEM(rootPem); !ok {
		return nil, errors.New("append root ca to cert pool failed")
	}
	cfg := &tls.Config{
		Rand:               rand.Reader,
		InsecureSkipVerify: false,
		CipherSuites:       DefaultSafeCipherSuites,
		MinVersion:         tls.VersionTLS13,
		RootCAs:            rootCaPool,
	}
	cfg.VerifyPeerCertificate = func(certificates [][]byte, verifiedChains [][]*x509.Certificate) error {
		for _, crl := range crls {
			if err := checkCaFileInCrl(verifiedChains, crl); err != nil {
				return err
			}
		}
		return nil
	}

	return cfg, nil
}

// for dynamic update service cert. cert pair will be loaded from disk on each time tls handshake
func getCertPairFromParam(params tlsParam) (*tls.Certificate, error) {
	pemPair, err := GetCertPairForPem(params.certPath, params.keyPath, params.KmcCfg)
	if err != nil {
		return nil, err
	}
	defer utils.ClearSliceByteMemory(pemPair.KeyPem)
	pair, err := tls.X509KeyPair(pemPair.CertPem, pemPair.KeyPem)
	if err != nil {
		return nil, err
	}
	return &pair, nil
}
