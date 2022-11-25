// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package certmanager constants used in cert manager module
package certmanager

// mefCerts
const (
	CertMgrPathName             = "mef_certs"
	RootCaNameValidCenter       = "cloud_root.crt" // PEM format
	RootCaBackUpNameValidCenter = "cloud_root_backup.crt"
	RootCaNameValidEdge         = "edge_root.crt"
	RootCaBackUpNameValidEdge   = "edge_root_backup.crt"
	ServiceCaName               = "edge_service.crt"
	KeyFileName                 = "edge.key"
	PwdFileName                 = "edge.pwd"

	CertDirMode  = 0700
	CertFileMode = 0400

	PrivateKeyBits  = 4096
	NewInt          = 2022
	ValidationYear  = 10
	ValidationMonth = 0
	ValidationDay   = 0

	ValidCenter = "validCenter"
	ValidEdge   = "validEdge"

	CaCountry            = "CN"
	CaOrganization       = "Huawei"
	CaOrganizationalUnit = "Ascend"
)

// CERTIFICATE pemType
const (
	CERTIFICATE = "CERTIFICATE"
)
