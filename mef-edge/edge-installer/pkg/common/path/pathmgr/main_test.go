// Copyright (c)  2024. Huawei Technologies Co., Ltd.  All rights reserved.

// Package pathmgr for package main test
package pathmgr

import (
	"testing"

	"huawei.com/mindx/common/test"
)

const (
	testInstallRootDir     = "/tmp"
	testInstallationPkgDir = "/installation"
	testLogRootDir         = "/tmp/log"
	testLogBackupRootDir   = "/tmp/log_backup"
)

var pathMgr *PathManager

func TestMain(m *testing.M) {
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, nil)
}
