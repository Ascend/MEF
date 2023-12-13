// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package handlers for package main test
package handlers

import (
	"testing"

	"huawei.com/mindx/common/test"
)

func TestMain(m *testing.M) {
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, nil)
}
