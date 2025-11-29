// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package innercommands for package main test
package innercommands

import (
	"testing"

	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/installer/edgectl/common"
)

var ctx = &common.Context{
	WorkPathMgr:   pathmgr.NewWorkPathMgr("./"),
	ConfigPathMgr: pathmgr.NewConfigPathMgr("./"),
}

func TestMain(m *testing.M) {
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, nil)
}
