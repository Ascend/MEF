// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package certconstant cert constant define
package certconstant

const (
	// RootCaPath root ca save directory
	RootCaPath = "/home/data/mef/root_ca"
	// RootCaFileName root ca save file name
	RootCaFileName = "root_ca.crt"
	// RootKeyFileName root key save file name
	RootKeyFileName = "encrypt_root_ca.key"
)

const (
	// ErrorGetRootCa query root ca failed
	ErrorGetRootCa = "00003001"
	// ErrorIssueSrvCert issue service certificate failed
	ErrorIssueSrvCert = "00003002"
)
