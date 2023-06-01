// Copyright (c) 2021. Huawei Technologies Co., Ltd. All rights reserved.

// Package certmanager cert manager module
package certmanager

import (
	"path/filepath"

	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

func getRootCaPath(certName string) string {
	return filepath.Join(util.RootCaMgrDir, certName, util.RootCaFileName)
}

func getRootKeyPath(certName string) string {
	return filepath.Join(util.RootCaMgrDir, certName, util.RootKeyFileName)
}
