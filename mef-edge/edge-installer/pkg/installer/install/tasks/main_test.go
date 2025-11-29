// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package tasks for package main test
package tasks

import (
	"path/filepath"
	"testing"

	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/path/pathmgr"
)

var (
	testDir        = "/tmp/test_install_tasks_dir"
	installDir     = filepath.Join(testDir, "mindx")
	softwareDir    = filepath.Join(testDir, "mindx/MEFEdge/software")
	logDir         = filepath.Join(testDir, "log")
	logBackupDir   = filepath.Join(testDir, "log_backup")
	originDir      = filepath.Join(testDir, "origin")
	pathMgr        = pathmgr.NewPathMgr(installDir, originDir, logDir, logBackupDir)
	workAbsPathMgr = pathmgr.NewWorkAbsPathMgr(softwareDir)
)

func TestMain(m *testing.M) {
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, nil)
}
