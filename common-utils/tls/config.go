// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tls provides the capability of tls.
package tls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"strings"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/rand"
	cx509 "huawei.com/mindx/common/x509"
)

var (
	dirPrefix = "/etc/mindx-dl/npu-exporter/"
)

const (
	prefix = ".config"
	// KeyStore KeyStore path
	KeyStore = prefix + "/config1"
	// KeyBackup key backup store file
	KeyBackup = ".config1"
	// CertStore CertStore path
	CertStore = prefix + "/config2"
	// CertBackup Cert Backup file
	CertBackup = ".config2"
	// CaStore CaStore path
	CaStore = prefix + "/config3"
	// CaBackup Ca Backup file
	CaBackup = ".config3"
	// CrlStore CrlStore path
	CrlStore = prefix + "/config4"
	// CrlBackup crl backup file
	CrlBackup = ".config4"
	// PassFile PassFile path
	PassFile = prefix + "/config5"
	// KubeCfgFile kubeconfig file store path
	KubeCfgFile = prefix + "/config6"
	// KubeCfgBackup kubeconfig backup file
	KubeCfgBackup = ".config6"
	// PassFileBackUp PassFileBackUp path
	PassFileBackUp = ".conf"
)

// NewTLSConfig  create the tls config struct
func NewTLSConfig(caBytes []byte, certificate tls.Certificate, cipherSuites []uint16) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: cipherSuites,
		Rand:         rand.Reader,
	}
	if len(caBytes) > 0 {
		// Two-way SSL
		pool := x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM(caBytes); !ok {
			return nil, errors.New("append the CA file failed")
		}
		tlsConfig.ClientCAs = pool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		hwlog.RunLog.Info("enable Two-way SSL mode")
	} else {
		// One-way SSL
		tlsConfig.ClientAuth = tls.NoClientCert
		hwlog.RunLog.Info("enable One-way SSL mode")
	}
	return tlsConfig, nil
}

// GetTLSConfigForClient get the tls config for client
func GetTLSConfigForClient(componentType string, encryptAlgorithm int) (*tls.Config, error) {
	if componentType != "npu-exporter" {
		dirPrefix = strings.Replace(dirPrefix, "npu-exporter", componentType, -1)
	}
	pathMap := map[string]string{
		cx509.CertStorePath:       dirPrefix + CertStore,
		cx509.CertStoreBackupPath: dirPrefix + CertBackup,
		cx509.KeyStorePath:        dirPrefix + KeyStore,
		cx509.KeyStoreBackupPath:  dirPrefix + KeyBackup,
		cx509.PassFilePath:        dirPrefix + PassFile,
		cx509.PassFileBackUpPath:  dirPrefix + PassFileBackUp,
	}
	tlsCert, err := LoadCertPair(pathMap, encryptAlgorithm)
	if err != nil {
		return nil, err
	}
	caInstance, err := cx509.NewBKPInstance(nil, dirPrefix+CaStore, dirPrefix+CaBackup)
	if err != nil {
		return nil, err
	}
	caBytes, err := caInstance.ReadFromDisk(fileutils.Mode600, false)
	if err != nil || caBytes == nil {
		hwlog.RunLog.Info("no ca file found")
		return nil, err
	}
	if err = cx509.VerifyCaCert(caBytes, cx509.InvalidNum); err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(caBytes); !ok {
		return nil, errors.New("append the CA file failed")
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*tlsCert},
		RootCAs:      pool,
		MinVersion:   tls.VersionTLS12,
		Rand:         rand.Reader,
	}
	return tlsConfig, nil
}
