// Copyright (c) 2021. Huawei Technologies Co., Ltd. All rights reserved.

// Package certmanager cert manager module
package certmanager

import (
	"path"

	"huawei.com/mindxedge/base/common"

	"cert-manager/pkg/certconstant"
)

var certImportMap = map[string]bool{
	common.WsSerName:        false,
	common.WsCltName:        false,
	common.SoftwareCertName: true,
	common.ImageCertName:    true,
	common.NginxCertName:    true,
	common.InnerName:        false,
}

// CheckCertName check use id if valid
func CheckCertName(certName string) bool {
	_, ok := certImportMap[certName]
	return ok
}

func checkIfCanImport(certName string) bool {
	return certImportMap[certName]
}

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
