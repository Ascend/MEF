// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package flows for main test
package flows

import (
	"testing"

	"huawei.com/mindx/common/test"
)

var testDir = "/tmp/test_preupgrade_flow"

func TestMain(m *testing.M) {
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, nil)
}
