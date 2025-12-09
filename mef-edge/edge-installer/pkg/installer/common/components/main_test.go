// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package components for main test
package components

import (
	"errors"
	"path/filepath"
	"testing"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
)

var (
	testDir         = "/tmp/test_components_dir"
	installDir      = filepath.Join(testDir, "mindx")
	softwareDir     = filepath.Join(testDir, "mindx/MEFEdge/software_A")
	logDir          = filepath.Join(testDir, "log")
	logBackupDir    = filepath.Join(testDir, "log_backup")
	originDir       = filepath.Join(testDir, "origin")
	pathMgr         = pathmgr.NewPathMgr(installDir, originDir, logDir, logBackupDir)
	workAbsPathMgr  = pathmgr.NewWorkAbsPathMgr(softwareDir)
	prepareCompBase = PrepareCompBase{PathManager: pathMgr, WorkAbsPathMgr: workAbsPathMgr}
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

	err1 := prepareCompPkgDir()
	err2 := prepareOtherFileOrDir()
	if err1 != nil || err2 != nil {
		hwlog.RunLog.Error("prepare dir or file failed")
		return errors.New("prepare dir or file failed")
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

func prepareCompPkgDir() error {
	dirNames := []string{constants.SoftwareDir, constants.Config}
	compNames := []string{constants.EdgeInstaller, constants.EdgeMain, constants.EdgeOm,
		constants.DevicePlugin, constants.EdgeCore, constants.Lib}
	for _, dirName := range dirNames {
		dir := filepath.Join(originDir, dirName)
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
		}
	}
	return nil
}

func prepareOtherFileOrDir() error {
	runShPath := filepath.Join(originDir, constants.SoftwareDir, constants.RunScript)
	if err := fileutils.CreateFile(runShPath, constants.Mode500); err != nil {
		hwlog.RunLog.Errorf("create file [%s] failed, error: %v", runShPath, err)
		return err
	}

	versionPath := filepath.Join(originDir, constants.VersionXml)
	if err := fileutils.CreateFile(versionPath, constants.Mode400); err != nil {
		hwlog.RunLog.Errorf("create file [%s] failed, error: %v", versionPath, err)
		return err
	}

	dirs := []string{
		softwareDir,
		prepareCompBase.PathManager.GetEdgeLogDir(),
		prepareCompBase.PathManager.GetEdgeLogBackupDir(),
		prepareCompBase.PathManager.ConfigPathMgr.GetConfigDir(),
	}
	for _, dir := range dirs {
		if err := fileutils.CreateDir(dir, constants.Mode755); err != nil {
			hwlog.RunLog.Errorf("create dir [%s] failed, error: %v", dir, err)
			return err
		}
	}
	return nil
}
