// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build TESTCODE
// +build TESTCODE

// Package testutils
package testutils

import (
	"errors"
	"fmt"
	"io"
	"os"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/test"

	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/constants"
)

// PrepareTempDirs prepares temp dirs
func PrepareTempDirs() error {
	if err := fileutils.DeleteAllFileWithConfusion("/home/MEFCenter"); err != nil {
		hwlog.RunLog.Errorf("delete /home/MEFCenter failed: %s", err.Error())
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

// TcLogMgr struct for test case base
type TcLogMgr struct{}

// Setup pre-processing
func (tc *TcLogMgr) Setup() error {
	if err := test.InitLog(); err != nil {
		return err
	}
	if err := PrepareTempDirs(); err != nil {
		fmt.Printf("prepare dirs failed, %v\n", err)
		return err
	}
	return nil
}

// Teardown post-processing
func (tc *TcLogMgr) Teardown() {
	if err := CleanupTempDirs(); err != nil {
		fmt.Printf("cleanup dirs failed, %v\n", err)
	}
}
