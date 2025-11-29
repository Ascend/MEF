// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package tasks for main test
package tasks

import (
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/installer/common/tasks"
)

var (
	testDir           = "/tmp/test_upgrade_tasks_dir"
	pathMgr           = pathmgr.NewPathMgr(testDir, testDir, testDir, testDir)
	setWorkPathTask   = SetWorkPathTask{PathMgr: pathMgr}
	postEffectProcess = PostEffectProcessTask{
		PostProcessBaseTask: tasks.PostProcessBaseTask{
			WorkPathMgr: pathMgr.SoftwarePathMgr.WorkPathMgr,
			LogPathMgr:  pathMgr.LogPathMgr,
		},
		ConfigPathMgr: pathMgr.ConfigPathMgr,
	}
)

func clearEnv(path string) {
	if err := fileutils.DeleteAllFileWithConfusion(path); err != nil {
		hwlog.RunLog.Errorf("clear env for test failed, error: %v", err)
		return
	}
}

func TestMain(m *testing.M) {
	tcModule := &test.TcBaseWithDb{DbPath: filepath.Join("/tmp", constants.DbEdgeMainPath)}
	test.RunWithPatches(tcModule, m, gomonkey.ApplyFunc(database.GetDb, test.MockGetDb))
}
