// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package certconstant cert constant define
package certconstant

const (
	// ServerCertPath  cert-manager server cert path
	ServerCertPath = "/home/data/config/mef-certs/cert-manager.crt"
	// ServerKeyPath cert-manager server key path
	ServerKeyPath = "/home/data/config/mef-certs/cert-manager.key"
	// RootCaPath  cert-manager server root ca path
	RootCaPath = "/home/data/inner-root-ca/RootCA.crt"
	// RootCaMgrDir root ca save directory
	RootCaMgrDir = "/home/data/config/root-ca/"
	// InnerRootCaDir inner root ca dir
	InnerRootCaDir = "/home/data/inner-root-ca/"
	// RootCaFileName root ca save file name
	RootCaFileName = "root.crt"
	// RootKeyFileName root key save file name
	RootKeyFileName = "encrypt_root.key"
)

const (
	// ErrorGetRootCa query root ca failed
	ErrorGetRootCa = "00003001"
	// ErrorIssueSrvCert issue service certificate failed
	ErrorIssueSrvCert = "00003002"
)
