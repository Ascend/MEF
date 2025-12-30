// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
