// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
