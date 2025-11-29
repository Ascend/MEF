// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common base file utils used
package common

import (
	"os"
)

// CreateSoftLink creates a softLink to dstPath on srcPath.
func CreateSoftLink(dstPath, srcPath string) error {
	return os.Symlink(dstPath, srcPath)
}
