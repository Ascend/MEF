// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package path for package main test
package path

import (
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"huawei.com/mindx/common/test"
)

func TestMain(m *testing.M) {
	patches := gomonkey.ApplyFuncReturn(filepath.EvalSymlinks, "./", nil)
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, patches)
}
