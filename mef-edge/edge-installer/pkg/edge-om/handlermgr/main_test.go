// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package handlermgr for package test main
package handlermgr

import (
	"errors"
	"testing"

	"huawei.com/mindx/common/test"
)

var testErr = errors.New("test error")

func TestMain(m *testing.M) {
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, nil)
}
