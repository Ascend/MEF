// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package flows for main test
package flows

import (
	"testing"

	"huawei.com/mindx/common/test"
)

func TestMain(m *testing.M) {
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, nil)
}
