// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_A500

// Package util
package util

// GetSerialNumber get serial number
func GetSerialNumber(installRootDir string) (string, error) {
	sn, err := getA500Sn()
	if err == nil {
		return sn, nil
	}
	return GetUuid()
}
