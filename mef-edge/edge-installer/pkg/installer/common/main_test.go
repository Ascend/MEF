// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package common for package test main
package common

import (
	"errors"
	"path/filepath"
	"testing"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
)

var (
	testErr      = errors.New("test error")
	testDir      = "/tmp/test_common_dir"
	logDir       = filepath.Join(testDir, "log")
	edgeDir      = filepath.Join(testDir, constants.MEFEdgeName)
	componentMgr = NewComponentMgr(testDir)
	compNames    = []string{constants.EdgeInstaller, constants.EdgeMain, constants.EdgeOm,
		constants.DevicePlugin, constants.EdgeCore}
)

type testCaseModule struct{}

// Setup pre-processing
func (tc *testCaseModule) Setup() error {
	if err := test.InitLog(); err != nil {
		return err
	}

	if err := fileutils.DeleteAllFileWithConfusion(testDir); err != nil {
		hwlog.RunLog.Errorf("cleanup [%s] failed, error: %v", testDir, err)
		return err
	}

	if err := prepareCompDir(); err != nil {
		hwlog.RunLog.Errorf("prepare dir failed, error: %v", err)
		return errors.New("prepare dir failed")
	}
	return nil
}

// Teardown post-processing
func (tc *testCaseModule) Teardown() {
	if err := fileutils.DeleteAllFileWithConfusion(testDir); err != nil {
		hwlog.RunLog.Warnf("cleanup [%s] failed, error: %v", testDir, err)
	}
}

func TestMain(m *testing.M) {
	test.RunWithPatches(&testCaseModule{}, m, nil)
}

func prepareCompDir() error {
	dirNames := []string{constants.SoftwareDir, constants.Config}
	testFile := "test.json"
	for _, dirName := range dirNames {
		dir := filepath.Join(edgeDir, dirName)
		if err := fileutils.CreateDir(dir, constants.Mode755); err != nil {
			hwlog.RunLog.Errorf("create origin dir [%s] failed, error: %v", dir, err)
			return err
		}

		for _, compName := range compNames {
			compDir := filepath.Join(dir, compName)
			if err := fileutils.CreateDir(compDir, constants.Mode700); err != nil {
				hwlog.RunLog.Errorf("create origin dir [%s] failed, error: %v", compDir, err)
				return err
			}
			if err := fileutils.CreateFile(filepath.Join(compDir, testFile), constants.Mode640); err != nil {
				hwlog.RunLog.Errorf("create test file [%s] failed, error: %v", testFile, err)
				return err
			}
		}
	}
	return nil
}
