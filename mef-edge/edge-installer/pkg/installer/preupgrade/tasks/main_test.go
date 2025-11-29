// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package tasks for main test
package tasks

import (
	"testing"

	"huawei.com/mindx/common/test"
)

var testDir = "/tmp/test_preupgrade_tasks"

func TestMain(m *testing.M) {
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, nil)
}
