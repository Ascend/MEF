// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package monitors for package main test
package monitors

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/test"

	"huawei.com/mindxedge/base/common"
)

func TestMain(m *testing.M) {
	patches := gomonkey.ApplyFuncReturn(common.GetHostIP, "", nil).
		ApplyFuncReturn(modulemgr.SendAsyncMessage, nil)
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, patches)
}
