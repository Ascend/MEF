// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common base file utils used
package common

import (
	"os"

	"huawei.com/mindx/common/envutils"
)

// CopyDir is used to copy dir and all files into it
func CopyDir(srcPath string, dstPath string, includeDir bool) error {
	if !includeDir {
		srcPath = srcPath + "/."
	}

	if _, err := envutils.RunCommand(CommandCopy, envutils.DefCmdTimeoutSec, "-r", srcPath, dstPath); err != nil {
		return err
	}
	return nil
}

// CreateSoftLink creates a softLink to dstPath on srcPath.
func CreateSoftLink(dstPath, srcPath string) error {
	return os.Symlink(dstPath, srcPath)
}
