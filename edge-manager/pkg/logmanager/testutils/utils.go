// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build TESTCODE
// +build TESTCODE

// Package testutils
package testutils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/constants"
)

// PrepareHwlog prepares hwlog
func PrepareHwlog() error {
	logConfig := &hwlog.LogConfig{OnlyToStdout: true}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		return err
	}
	return hwlog.InitOperateLogger(logConfig, context.Background())
}

// PrepareTempDirs prepares temp dirs
func PrepareTempDirs() error {
	if err := fileutils.DeleteAllFileWithConfusion("/home/MEFCenter"); err != nil {
		return fmt.Errorf("delete /home/MEFCenter failed: %s", err.Error())
	}
	dirs := []string{constants.LogDumpTempDir, constants.LogDumpPublicDir}
	for _, dir := range dirs {
		if err := fileutils.DeleteAllFileWithConfusion(dir); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err := fileutils.CreateDir(dir, common.Mode700); err != nil {
			return err
		}
	}
	return nil
}

// CleanupTempDirs cleans up temp dirs
func CleanupTempDirs() error {
	dirs := []string{constants.LogDumpTempDir, constants.LogDumpPublicDir}
	for _, dir := range dirs {
		if err := fileutils.DeleteAllFileWithConfusion(dir); err != nil && errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

// WithoutDiskPressureProtect returns a writer
func WithoutDiskPressureProtect(writer io.Writer, filePath string) io.Writer {
	return writer
}
