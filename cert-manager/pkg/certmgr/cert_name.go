// Copyright (c) 2021. Huawei Technologies Co., Ltd. All rights reserved.

// Package certmgr  cert mgr
package certmgr

import (
	"path"

	"cert-manager/pkg/certconstant"
	"huawei.com/mindxedge/base/common"
)

var certImportMap = map[string]bool{
	common.InnerName:            false,
	common.WsSerName:            false,
	common.WsCltName:            false,
	common.SoftwareCertName:     true,
	common.ImageCertName:        true,
	common.ResFileCertName:      true,
	common.NginxCertName:        true,
	common.EdgeCoreCertName:     true,
	common.DevicePluginCertName: true,
	common.AlarmCertName:        true,
}

// checkCertName check use id if valid
func checkCertName(certName string) bool {
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
	return path.Join(certconstant.InnerRootCaDir, certconstant.RootCaFileName)
}

func getInnerRootKeyPath() string {
	return path.Join(certconstant.InnerRootCaDir, certconstant.RootKeyFileName)
}
