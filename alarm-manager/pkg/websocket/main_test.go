// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package websocket for package main test
package websocket

import (
	"testing"

	"huawei.com/mindx/common/test"
)

func TestMain(m *testing.M) {
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, nil)
}
