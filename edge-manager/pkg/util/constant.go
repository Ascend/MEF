// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package util for edge-manager
package util

const (
	// ServerCertPath server cert path
	ServerCertPath = "/home/data/config/mef-certs/edge-manager.crt"
	// ServerKeyPath server encrypt key path
	ServerKeyPath = "/home/data/config/mef-certs/edge-manager.key"
	// RootCaPath root ca path
	RootCaPath = "/home/data/inner-root-ca/RootCA.crt"
	// CloudCoreCaPath cloud core root ca path
	CloudCoreCaPath = "/home/data/config/cloud-core-certs/rootCA.crt"
)

const (
	// ConnCheckUrl for mef edge do test connection
	ConnCheckUrl = "/check/conn"
)
