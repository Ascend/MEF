// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package utils for package main test
package utils

import (
	"testing"

	"huawei.com/mindx/common/test"

	"edge-manager/pkg/logmanager/testutils"
)

func TestMain(m *testing.M) {
	tcLogMgr := &testutils.TcLogMgr{}
	test.RunWithPatches(tcLogMgr, m, nil)
}
