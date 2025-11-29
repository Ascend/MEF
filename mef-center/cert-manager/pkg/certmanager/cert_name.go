// Copyright (c) 2021. Huawei Technologies Co., Ltd. All rights reserved.

// Package certmanager cert manager module
package certmanager

import (
	"path/filepath"

	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

const (
	tempFileSuffix = ".tmp"
)

func getRootCaDir(certName string) string {
	return filepath.Join(util.RootCaMgrDir, certName)
}

func getRootCaPath(certName string) string {
	return filepath.Join(getRootCaDir(certName), util.RootCaFileName)
}

func getRootKeyPath(certName string) string {
	return filepath.Join(getRootCaDir(certName), util.RootKeyFileName)
}

func getCrlPath(crlName string) string {
	return filepath.Join(util.RootCaMgrDir, crlName, util.CrlName)
}

func getTempRootCaPath(certName string) string {
	return filepath.Join(getRootCaDir(certName), util.RootCaFileName+tempFileSuffix)
}

func getTempRootKeyPath(certName string) string {
	return filepath.Join(getRootCaDir(certName), util.RootKeyFileName+tempFileSuffix)
}
