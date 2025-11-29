// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_SDK

package util

import (
	"path/filepath"

	"huawei.com/mindx/common/fileutils"

	"edge-installer/pkg/common/constants"
)

// DeleteImageCertFile delete image cert file
func DeleteImageCertFile(imageAddress string) error {
	imageCertPath := filepath.Join(constants.DockerCertDir, imageAddress)
	return fileutils.DeleteAllFileWithConfusion(imageCertPath)
}
