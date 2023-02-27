// Copyright (c) 2021. Huawei Technologies Co., Ltd. All rights reserved.

// Package certmanager cert manager module
package certmanager

import (
	"path"

	"cert-manager/pkg/certconstant"
)

func getRootCaPath(certName string) string {
	return path.Join(certconstant.RootCaMgrDir, certName, certconstant.RootCaFileName)
}

func getRootKeyPath(certName string) string {
	return path.Join(certconstant.RootCaMgrDir, certName, certconstant.RootKeyFileName)
}

func getInnerRootCaPath() string {
	return path.Join(certconstant.InnerRootCaDir, certconstant.InnerCaFileName)
}

func getInnerRootKeyPath() string {
	return path.Join(certconstant.InnerRootCaDir, certconstant.RootKeyFileName)
}
